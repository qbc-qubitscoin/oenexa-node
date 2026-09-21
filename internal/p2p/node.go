package p2p

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"

	oenexaCrypto "github.com/oenexa/oenexa/internal/crypto"
	"github.com/oenexa/oenexa/internal/core"
)

const (
	TxTopic             = "/oenexa/tx/1.0.0"
	BlockTopic          = "/oenexa/block/1.0.0"
	VoteTopic           = "/oenexa/vote/1.0.0"
	SyncReqTopic        = "/oenexa/syncreq/1.0.0"
	SyncResTopic        = "/oenexa/syncres/1.0.0"
	DiscoveryServiceTag = "oenexa-network-discovery"
)

// Sync structs for pubsub
type SyncReq struct {
	FromHeight uint64 `json:"from_height"`
	MaxCount   uint32 `json:"max_count"`
}

type SyncRes struct {
	Blocks []*core.Block `json:"blocks"`
}

// Node is the libp2p network participant.
type Node struct {
	host     host.Host
	pubsub   *pubsub.PubSub
	identity *Identity

	txTopic      *pubsub.Topic
	blockTopic   *pubsub.Topic
	voteTopic    *pubsub.Topic
	syncReqTopic *pubsub.Topic
	syncResTopic *pubsub.Topic

	txSub      *pubsub.Subscription
	blockSub   *pubsub.Subscription
	voteSub    *pubsub.Subscription
	syncReqSub *pubsub.Subscription
	syncResSub *pubsub.Subscription

	// Hooks
	OnTxReceived    func(*core.Transaction)
	OnBlockReceived func(*core.Block)
	OnVoteReceived  func([]byte)
	OnSyncReq       func(SyncReq)
	OnSyncRes       func(SyncRes)
}

// mdnsNotifee handles mDNS peer discovery events.
type mdnsNotifee struct {
	h host.Host
}

func (n *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		return
	}
	log.Printf("[p2p] Discovered peer via mDNS: %s", pi.ID.String())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		log.Printf("[p2p] Failed to connect to discovered peer: %v", err)
	} else {
		log.Printf("[p2p] Connected to %s", pi.ID.String())
	}
}

// NewNode creates a libp2p node.
func NewNode(local *Identity, listenPort int) (*Node, error) {
	// Generate a deterministic libp2p Ed25519 key from the ML-DSA-65 private key hash
	// so the peer ID is consistent across restarts.
	seed := oenexaCrypto.Hash256(local.PrivateKey)
	edKey := ed25519.NewKeyFromSeed(seed[:32])
	
	privKey, err := crypto.UnmarshalEd25519PrivateKey(edKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate libp2p key: %w", err)
	}

	listenAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)
	
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(listenAddr),
		libp2p.Identity(privKey),
		libp2p.DefaultTransports,
		libp2p.DefaultSecurity,
		libp2p.DefaultMuxers,
	)
	if err != nil {
		return nil, fmt.Errorf("libp2p.New: %w", err)
	}

	ps, err := pubsub.NewGossipSub(context.Background(), h)
	if err != nil {
		return nil, fmt.Errorf("pubsub.NewGossipSub: %w", err)
	}

	return &Node{
		host:     h,
		pubsub:   ps,
		identity: local,
	}, nil
}

// Start begins gossiping and listening for incoming pubsub messages.
func (n *Node) Start(ctx context.Context) error {
	log.Printf("[p2p] libp2p node started. ID: %s", n.host.ID())
	for _, addr := range n.host.Addrs() {
		log.Printf("[p2p] Listening on: %s/p2p/%s", addr, n.host.ID())
	}

	// Setup mDNS discovery
	mdnsService := mdns.NewMdnsService(n.host, DiscoveryServiceTag, &mdnsNotifee{h: n.host})
	if err := mdnsService.Start(); err != nil {
		log.Printf("[p2p] warning: mDNS discovery failed to start: %v", err)
	}

	var err error
	// Join topics
	n.txTopic, err = n.pubsub.Join(TxTopic)
	if err != nil {
		return err
	}
	n.blockTopic, err = n.pubsub.Join(BlockTopic)
	if err != nil {
		return err
	}
	n.voteTopic, err = n.pubsub.Join(VoteTopic)
	if err != nil {
		return err
	}
	n.syncReqTopic, err = n.pubsub.Join(SyncReqTopic)
	if err != nil {
		return err
	}
	n.syncResTopic, err = n.pubsub.Join(SyncResTopic)
	if err != nil {
		return err
	}

	// Subscribe
	n.txSub, err = n.txTopic.Subscribe()
	if err != nil {
		return err
	}
	n.blockSub, err = n.blockTopic.Subscribe()
	if err != nil {
		return err
	}
	n.voteSub, err = n.voteTopic.Subscribe()
	if err != nil {
		return err
	}
	n.syncReqSub, err = n.syncReqTopic.Subscribe()
	if err != nil {
		return err
	}
	n.syncResSub, err = n.syncResTopic.Subscribe()
	if err != nil {
		return err
	}

	// Run listener loops
	go n.handleTxSub(ctx)
	go n.handleBlockSub(ctx)
	go n.handleVoteSub(ctx)
	go n.handleSyncReqSub(ctx)
	go n.handleSyncResSub(ctx)

	return nil
}

// BroadcastSyncReq gossips a sync request to the network.
func (n *Node) BroadcastSyncReq(ctx context.Context, req SyncReq) error {
	if n.syncReqTopic == nil {
		return nil
	}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return n.syncReqTopic.Publish(ctx, data)
}

// BroadcastSyncRes gossips a sync response to the network.
func (n *Node) BroadcastSyncRes(ctx context.Context, res SyncRes) error {
	if n.syncResTopic == nil {
		return nil
	}
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	return n.syncResTopic.Publish(ctx, data)
}

// Connect dials an explicit multiaddress.
func (n *Node) Connect(ctx context.Context, multiaddr string) error {
	addrInfo, err := peer.AddrInfoFromString(multiaddr)
	if err != nil {
		return err
	}
	return n.host.Connect(ctx, *addrInfo)
}

// PeerCount returns the number of connected libp2p peers.
func (n *Node) PeerCount() int {
	return len(n.host.Network().Peers())
}

// ListenAddrs returns the node's listening multiaddresses.
func (n *Node) ListenAddrs() []string {
	var addrs []string
	for _, addr := range n.host.Addrs() {
		addrs = append(addrs, fmt.Sprintf("%s/p2p/%s", addr.String(), n.host.ID().String()))
	}
	return addrs
}

// BroadcastTx gossips a transaction to the network.
func (n *Node) BroadcastTx(ctx context.Context, tx *core.Transaction) error {
	if n.txTopic == nil {
		return nil
	}
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return n.txTopic.Publish(ctx, data)
}

// BroadcastBlock gossips a block to the network.
func (n *Node) BroadcastBlock(ctx context.Context, blk *core.Block) error {
	if n.blockTopic == nil {
		return nil
	}
	data, err := json.Marshal(blk)
	if err != nil {
		return err
	}
	return n.blockTopic.Publish(ctx, data)
}

func (n *Node) handleTxSub(ctx context.Context) {
	for {
		msg, err := n.txSub.Next(ctx)
		if err != nil {
			return
		}
		// Skip our own messages
		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		var tx core.Transaction
		if err := json.Unmarshal(msg.Data, &tx); err != nil {
			log.Printf("[p2p] failed to unmarshal tx: %v", err)
			continue
		}

		if n.OnTxReceived != nil {
			n.OnTxReceived(&tx)
		}
	}
}

func (n *Node) handleBlockSub(ctx context.Context) {
	for {
		msg, err := n.blockSub.Next(ctx)
		if err != nil {
			return
		}
		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		var block core.Block
		if err := json.Unmarshal(msg.Data, &block); err != nil {
			log.Printf("[p2p] failed to unmarshal block: %v", err)
			continue
		}

		if n.OnBlockReceived != nil {
			n.OnBlockReceived(&block)
		}
	}
}

func (n *Node) handleVoteSub(ctx context.Context) {
	for {
		msg, err := n.voteSub.Next(ctx)
		if err != nil {
			return
		}
		if msg.ReceivedFrom == n.host.ID() {
			continue
		}
		if n.OnVoteReceived != nil {
			n.OnVoteReceived(msg.Data)
		}
	}
}

// BroadcastVote gossips a consensus vote.
func (n *Node) BroadcastVote(ctx context.Context, voteData []byte) error {
	if n.voteTopic == nil {
		return nil
	}
	return n.voteTopic.Publish(ctx, voteData)
}

func (n *Node) handleSyncReqSub(ctx context.Context) {
	for {
		msg, err := n.syncReqSub.Next(ctx)
		if err != nil {
			return
		}
		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		var req SyncReq
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			continue
		}

		if n.OnSyncReq != nil {
			n.OnSyncReq(req)
		}
	}
}

func (n *Node) handleSyncResSub(ctx context.Context) {
	for {
		msg, err := n.syncResSub.Next(ctx)
		if err != nil {
			return
		}
		if msg.ReceivedFrom == n.host.ID() {
			continue
		}

		var res SyncRes
		if err := json.Unmarshal(msg.Data, &res); err != nil {
			continue
		}

		if n.OnSyncRes != nil {
			n.OnSyncRes(res)
		}
	}
}

