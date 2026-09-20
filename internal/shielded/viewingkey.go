package shielded

import (
	"bytes"
	"encoding/binary"
	"errors"

	"golang.org/x/crypto/sha3"
)

// ViewingKey represents a read-only audit key that allows selective inspection of notes.
type ViewingKey [32]byte

// DeriveViewingKey creates a viewing key from a spending private key:
// vk = SHA-3-256("OENEXA_VIEWING_KEY" || spendingKey)
func DeriveViewingKey(spendingKey []byte) ViewingKey {
	h := sha3.New256()
	h.Write([]byte("OENEXA_VIEWING_KEY"))
	h.Write(spendingKey)
	var vk ViewingKey
	copy(vk[:], h.Sum(nil))
	return vk
}

// PaymentDisclosureProof is an exportable cryptographic proof verifying a specific shielded payment.
type PaymentDisclosureProof struct {
	TxHash      [32]byte
	Value       uint64
	RecipientPK []byte
	Commitment  [32]byte
	ProofSig    [32]byte // SHA-3-256("OENEXA_PAYMENT_PROOF" || vk || txHash || cm || value)
}

// CreatePaymentProof generates a proof that a payment was made to a recipient.
func CreatePaymentProof(note *Note, spendingKey []byte, txHash [32]byte) (*PaymentDisclosureProof, error) {
	if note == nil {
		return nil, errors.New("shielded: nil note")
	}
	if len(spendingKey) == 0 {
		return nil, errors.New("shielded: empty spending key")
	}

	vk := DeriveViewingKey(spendingKey)
	cm := note.Commitment()

	h := sha3.New256()
	h.Write([]byte("OENEXA_PAYMENT_PROOF"))
	h.Write(vk[:])
	h.Write(txHash[:])
	h.Write(cm[:])
	var b8 [8]byte
	binary.BigEndian.PutUint64(b8[:], note.Value)
	h.Write(b8[:])
	var proofSig [32]byte
	copy(proofSig[:], h.Sum(nil))

	return &PaymentDisclosureProof{
		TxHash:      txHash,
		Value:       note.Value,
		RecipientPK: note.RecipientPK,
		Commitment:  cm,
		ProofSig:    proofSig,
	}, nil
}

// VerifyPaymentProof verifies that a payment proof matches the expected transaction, recipient, and value.
func VerifyPaymentProof(proof *PaymentDisclosureProof, vk ViewingKey, expectedTxHash [32]byte, expectedRecipientPK []byte, expectedValue uint64) bool {
	if proof == nil {
		return false
	}
	if proof.TxHash != expectedTxHash {
		return false
	}
	if proof.Value != expectedValue {
		return false
	}
	if !bytes.Equal(proof.RecipientPK, expectedRecipientPK) {
		return false
	}

	h := sha3.New256()
	h.Write([]byte("OENEXA_PAYMENT_PROOF"))
	h.Write(vk[:])
	h.Write(expectedTxHash[:])
	h.Write(proof.Commitment[:])
	var b8 [8]byte
	binary.BigEndian.PutUint64(b8[:], expectedValue)
	h.Write(b8[:])
	var expectedSig [32]byte
	copy(expectedSig[:], h.Sum(nil))

	return proof.ProofSig == expectedSig
}
