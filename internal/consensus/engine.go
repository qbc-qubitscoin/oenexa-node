package consensus

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/oenexa/oenexa/internal/core"
	"github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/mempool"
	"github.com/oenexa/oenexa/internal/state"
	"github.com/oenexa/oenexa/internal/upgrade"
	"github.com/oenexa/oenexa/internal/vm"
)

const (
	BlockInterval = 2 * time.Second
	RoundTimeout  = 4 * time.Second
	MaxTxPerBlock = 10_000 // raised to match the higher block gas limit
)

type BFTStep int

const (
	StepPropose BFTStep = iota
	StepPrevote
	StepPrecommit
	StepCommit
)

// Engine drives BFT block production and validation.
type Engine struct {
	mu sync.Mutex

	validatorAddr [crypto.AddressSize]byte
	validatorPub  []byte
	validatorPriv []byte

	validatorSet *ValidatorSet
	state        *state.DB
	pool         *mempool.Mempool
	execVM       *vm.VM
	upgradeMgr   *upgrade.Manager

	chain    []*core.Block
	commitCh chan *core.Block

	// BFT State
	step          BFTStep
	round         uint32
	proposal      *core.Block
	proposalState *state.DB
	prevotes      map[[crypto.HashSize]byte]map[string]*Vote // blockHash -> validator -> Vote
	precommits    map[[crypto.HashSize]byte]map[string]*Vote

	// P2P Hooks
	OnVoteBroadcast  func(*Vote)
	OnBlockBroadcast func(*core.Block)
}

// NewEngine creates a consensus engine.
// execVM and upgradeMgr may be nil if those features are not needed.
func NewEngine(
	addr [crypto.AddressSize]byte,
	pubKey, privKey []byte,
	vs *ValidatorSet,
	st *state.DB,
	pool *mempool.Mempool,
	genesis *core.Block,
	execVM *vm.VM,
	upgradeMgr *upgrade.Manager,
) *Engine {
	return &Engine{
		validatorAddr: addr,
		validatorPub:  pubKey,
		validatorPriv: privKey,
		validatorSet:  vs,
		state:         st,
		pool:          pool,
		execVM:        execVM,
		upgradeMgr:    upgradeMgr,
		chain:         []*core.Block{genesis},
		commitCh:      make(chan *core.Block, 64),
		prevotes:      make(map[[crypto.HashSize]byte]map[string]*Vote),
		precommits:    make(map[[crypto.HashSize]byte]map[string]*Vote),
	}
}

var (
	blockInterval = BlockInterval
	applyTxFunc   = state.ApplyTransaction
	newBlockFunc  = core.NewBlock
	commitFunc    = func(e *Engine, blk *core.Block, snap *state.DB) error {
		return e.commit(blk, snap)
	}
)

// Run starts the block-production loop; it blocks until ctx is canceled.
func (e *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(blockInterval)
	defer ticker.Stop()
	log.Printf("[consensus] engine started, validator=%s", crypto.ToHex(e.validatorAddr))
	for {
		select {
		case <-ctx.Done():
			log.Println("[consensus] engine stopped")
			return
		case <-ticker.C:
			if err := e.produceBlock(ctx); err != nil {
				log.Printf("[consensus] block production error: %v", err)
			}
		}
	}
}

func (e *Engine) produceBlock(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	height := uint64(len(e.chain))
	proposer := e.validatorSet.Proposer(height)
	if proposer.Address != e.validatorAddr {
		return nil // not our turn
	}

	if e.step != StepPropose {
		return nil // already proposing/voting for this height
	}

	blk, newState, err := e.buildBlock(height)
	if err != nil {
		return fmt.Errorf("buildBlock: %w", err)
	}
	
	e.proposal = blk
	e.proposalState = newState
	e.step = StepPrevote

	log.Printf("[consensus] Proposing block %d (%x)", height, blk.Hash[:4])

	if e.OnBlockBroadcast != nil {
		e.OnBlockBroadcast(blk)
	}

	// Since we proposed it, we also prevote for it
	e.castVoteLocked(VotePrevote, height, blk.Hash)
	return nil
}

// ProcessProposal receives a block proposal from the network.
func (e *Engine) ProcessProposal(blk *core.Block) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	height := uint64(len(e.chain))
	if blk.Header.Height != height {
		return fmt.Errorf("proposal height %d does not match chain height %d", blk.Header.Height, height)
	}

	proposer := e.validatorSet.Proposer(height)
	if proposer == nil || blk.Header.ValidatorAddr != proposer.Address {
		return fmt.Errorf("invalid proposer")
	}

	// 1. Cryptographic validation: Verify proposer's ML-DSA-65 header signature
	if len(proposer.PublicKey) > 0 {
		if err := blk.VerifyValidatorSig(proposer.PublicKey); err != nil {
			return fmt.Errorf("invalid block validator signature: %w", err)
		}
	}

	// 2. Chain linkage validation
	prev := e.chain[height-1]
	if blk.Header.PrevHash != prev.Hash {
		return fmt.Errorf("invalid parent block hash")
	}
	if blk.Header.Timestamp <= prev.Header.Timestamp {
		return fmt.Errorf("block timestamp %d not after parent timestamp %d", blk.Header.Timestamp, prev.Header.Timestamp)
	}

	// 3. Merkle root validation of transactions
	txHashes := make([][crypto.HashSize]byte, len(blk.Txs))
	for i, tx := range blk.Txs {
		txHashes[i] = tx.Hash
	}
	expectedMerkle := core.ComputeMerkleRoot(txHashes)
	if expectedMerkle != blk.Header.MerkleRoot {
		return fmt.Errorf("invalid transactions merkle root")
	}

	// 4. State transition replay and validation
	snap := e.state.Snapshot()
	baseFee := prev.Header.BaseFee
	var totalGas, totalBurned, totalTip uint64

	for _, tx := range blk.Txs {
		result, err := applyTxFunc(snap, tx, core.BlockGasLimit-totalGas, e.execVM, baseFee)
		if err != nil {
			return fmt.Errorf("transaction execution failed (%s): %w", crypto.ToHex(tx.Hash), err)
		}
		totalGas += result.GasUsed
		totalBurned += result.BurnedFee
		totalTip += result.ValidatorTip
		if totalGas > core.BlockGasLimit {
			return fmt.Errorf("block gas limit exceeded")
		}
	}

	if totalGas != blk.Header.GasUsed {
		return fmt.Errorf("gas used mismatch: header has %d, computed %d", blk.Header.GasUsed, totalGas)
	}
	if totalBurned != blk.Header.BurnedFees {
		return fmt.Errorf("burned fees mismatch: header has %d, computed %d", blk.Header.BurnedFees, totalBurned)
	}

	// Credit validator: tip + block subsidy
	reward := core.BlockReward(height)
	income := totalTip + reward
	if income > 0 {
		valAcc := snap.GetAccount(proposer.Address)
		valAcc.Balance += income
		snap.SetAccount(proposer.Address, valAcc)
	}

	computedRoot := snap.CommitRoot()
	if computedRoot != blk.Header.StateRoot {
		return fmt.Errorf("state root mismatch: header has %s, computed %s", crypto.ToHex(blk.Header.StateRoot), crypto.ToHex(computedRoot))
	}

	if e.step > StepPropose {
		return nil // already processing or past propose step
	}

	e.proposal = blk
	e.proposalState = snap
	e.step = StepPrevote

	// Broadcast our prevote
	e.castVoteLocked(VotePrevote, height, blk.Hash)
	return nil
}

// ProcessVote receives a BFT vote from the network.
func (e *Engine) ProcessVote(v *Vote) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := v.Verify(); err != nil {
		return err
	}

	height := uint64(len(e.chain))
	if v.Height != height {
		return nil
	}

	if !e.validatorSet.Contains(v.Voter) {
		return fmt.Errorf("vote from non-validator")
	}

	voterHex := crypto.ToHex(v.Voter)
	hash := v.BlockHash

	// Equivocation detection: check if validator voted for a different hash at the same height & round
	votesMap := e.prevotes
	if v.Type == VotePrecommit {
		votesMap = e.precommits
	}
	for otherHash, voterVotes := range votesMap {
		if otherHash != hash {
			if existingVote, exists := voterVotes[voterHex]; exists && existingVote.Round == v.Round {
				log.Printf("[consensus] SECURITY ALERT: Equivocation detected from validator %s at height %d round %d! Slashing validator.", voterHex, v.Height, v.Round)
				e.slashValidatorLocked(v.Voter)
				return fmt.Errorf("equivocation detected from validator %s", voterHex)
			}
		}
	}

	if v.Type == VotePrevote {
		if e.prevotes[hash] == nil {
			e.prevotes[hash] = make(map[string]*Vote)
		}
		e.prevotes[hash][voterHex] = v
		e.checkPrevoteQuorumLocked(height, hash)
	} else if v.Type == VotePrecommit {
		if e.precommits[hash] == nil {
			e.precommits[hash] = make(map[string]*Vote)
		}
		e.precommits[hash][voterHex] = v
		e.checkPrecommitQuorumLocked(height, hash)
	}
	return nil
}

func (e *Engine) slashValidatorLocked(addr [crypto.AddressSize]byte) {
	// 1. Slash voting power in validator set
	e.validatorSet.Slash(addr, 1)

	// 2. Slash 10% on-chain balance in state if account exists
	if e.state != nil {
		acc := e.state.GetAccount(addr)
		if acc != nil && acc.Balance > 0 {
			penalty := acc.Balance / 10
			acc.Balance -= penalty
			e.state.SetAccount(addr, acc)
			log.Printf("[consensus] Slashed %d OEN from validator %s", penalty/core.OneOEN, crypto.ToHex(addr))
		}
	}
}

func (e *Engine) castVoteLocked(voteType uint8, height uint64, hash [crypto.HashSize]byte) {
	v := &Vote{
		Type:      voteType,
		Height:    height,
		Round:     e.round,
		BlockHash: hash,
		Voter:     e.validatorAddr,
		PublicKey: e.validatorPub,
	}
	_ = v.Sign(e.validatorPriv)

	if voteType == VotePrevote {
		if e.prevotes[hash] == nil {
			e.prevotes[hash] = make(map[string]*Vote)
		}
		e.prevotes[hash][crypto.ToHex(e.validatorAddr)] = v
	} else if voteType == VotePrecommit {
		if e.precommits[hash] == nil {
			e.precommits[hash] = make(map[string]*Vote)
		}
		e.precommits[hash][crypto.ToHex(e.validatorAddr)] = v
	}

	if e.OnVoteBroadcast != nil {
		e.OnVoteBroadcast(v)
	}

	if voteType == VotePrevote {
		e.checkPrevoteQuorumLocked(height, hash)
	} else {
		e.checkPrecommitQuorumLocked(height, hash)
	}
}

func (e *Engine) checkPrevoteQuorumLocked(height uint64, hash [crypto.HashSize]byte) {
	if e.step >= StepPrecommit {
		return
	}
	var voters [][crypto.AddressSize]byte
	for _, vote := range e.prevotes[hash] {
		voters = append(voters, vote.Voter)
	}
	if e.validatorSet.HasQuorum(voters) {
		e.step = StepPrecommit
		e.castVoteLocked(VotePrecommit, height, hash)
	}
}

func (e *Engine) checkPrecommitQuorumLocked(height uint64, hash [crypto.HashSize]byte) {
	if e.step >= StepCommit {
		return
	}
	var voters [][crypto.AddressSize]byte
	for _, vote := range e.precommits[hash] {
		voters = append(voters, vote.Voter)
	}
	if e.validatorSet.HasQuorum(voters) {
		e.step = StepCommit
		log.Printf("[consensus] Quorum reached for block %d. Committing.", height)
		
		// If we are a follower, we might not have `e.proposalState`. We'd apply it here.
		// For prototype simplicity, if we lack it, we build it.
		stateToCommit := e.proposalState
		if stateToCommit == nil && e.proposal != nil {
			// Fast forward validation build
			_, stateToCommit, _ = e.buildBlock(height)
		}

		if e.proposal != nil && stateToCommit != nil {
			_ = commitFunc(e, e.proposal, stateToCommit)
		}
		
		// Reset for next height
		e.step = StepPropose
		e.round = 0
		e.proposal = nil
		e.proposalState = nil
		e.prevotes = make(map[[crypto.HashSize]byte]map[string]*Vote)
		e.precommits = make(map[[crypto.HashSize]byte]map[string]*Vote)
	}
}

// buildBlock assembles a new block on a state snapshot.
// Fee model: baseFee is burned; only priorityTip + blockReward go to validator.
func (e *Engine) buildBlock(height uint64) (*core.Block, *state.DB, error) {
	prev := e.chain[height-1]
	prevHash := prev.Hash
	baseFee := prev.Header.BaseFee // inherit parent's base fee
	snap := e.state.Snapshot()

	pending := e.pool.Pending(MaxTxPerBlock)
	var included []*core.Transaction
	var totalGas, totalBurned, totalTip uint64

	for _, tx := range pending {
		result, err := applyTxFunc(snap, tx, core.BlockGasLimit-totalGas, e.execVM, baseFee)
		if err != nil {
			continue // skip txs that can't pay baseFee or are otherwise invalid
		}
		totalGas += result.GasUsed
		totalBurned += result.BurnedFee
		totalTip += result.ValidatorTip
		included = append(included, tx)
		if totalGas >= core.BlockGasLimit {
			break
		}
	}

	// ── Credit validator: tip + block subsidy (NOT burned fees) ─────────────
	reward := core.BlockReward(height)
	income := totalTip + reward
	if income > 0 {
		validator := snap.GetAccount(e.validatorAddr)
		validator.Balance += income
		snap.SetAccount(e.validatorAddr, validator)
	}

	stateRoot := snap.CommitRoot()

	// ── Compute next block's base fee ────────────────────────────────────────
	nextBaseFee := core.NextBaseFee(baseFee, totalGas)

	blk, err := newBlockFunc(
		height, prevHash, stateRoot,
		time.Now().UnixNano(),
		e.validatorAddr, included, totalGas, nextBaseFee, totalBurned,
	)
	if err != nil {
		return nil, nil, err
	}
	if err := blk.SignHeader(e.validatorPriv); err != nil {
		return nil, nil, err
	}
	return blk, snap, nil
}

// commit finalizes a block: applies state, purges mempool, appends a chain.
func (e *Engine) commit(blk *core.Block, snap *state.DB) error {
	e.state.Apply(snap)
	e.pool.PurgeCommitted(blk.Txs)
	e.chain = append(e.chain, blk)

	reward := core.BlockReward(blk.Header.Height)
	log.Printf("[consensus] block height=%d hash=%s txs=%d gas=%d baseFee=%d burned=%d oenexa reward=%d OEN",
		blk.Header.Height,
		crypto.ToHex(blk.Hash)[:16]+"…",
		len(blk.Txs),
		blk.Header.GasUsed,
		blk.Header.BaseFee,
		blk.Header.BurnedFees,
		reward/core.OneOEN,
	)

	select {
	case e.commitCh <- blk:
	default:
	}

	if e.upgradeMgr != nil {
		e.upgradeMgr.OnBlock(blk.Header.Height)
	}
	return nil
}

// CommitCh returns a channel that receives every committed block.
func (e *Engine) CommitCh() <-chan *core.Block { return e.commitCh }

// InjectBlock appends a block that was loaded from persistent storage during
// node startup.  It does NOT re-apply state — the caller must ensure state
// is already consistent with the injected blocks.
func (e *Engine) InjectBlock(blk *core.Block) {
	e.mu.Lock()
	defer e.mu.Unlock()
	// Extend the chain slice to the required height, filling gaps with nil.
	for uint64(len(e.chain)) <= blk.Header.Height {
		e.chain = append(e.chain, nil)
	}
	e.chain[blk.Header.Height] = blk
}

// BlockByHeight returns the block at the given height or nil.
func (e *Engine) BlockByHeight(h uint64) *core.Block {
	e.mu.Lock()
	defer e.mu.Unlock()
	if h >= uint64(len(e.chain)) {
		return nil
	}
	return e.chain[h]
}

// Height returns the current chain height (number of blocks including genesis).
func (e *Engine) Height() uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return uint64(len(e.chain))
}

// AcceptVote processes an incoming vote received from a peer validator.
// Extension point: not exercised in single-validator mode; called by the
// P2P layer when vote messages arrive from other validators.
func (e *Engine) AcceptVote(v *Vote) error {
	if err := v.Verify(); err != nil {
		return err
	}
	if !e.validatorSet.Contains(v.Voter) {
		return fmt.Errorf("unknown voter: %s", crypto.ToHex(v.Voter))
	}
	return nil
}
