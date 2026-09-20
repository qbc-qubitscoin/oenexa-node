package shielded

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/cloudflare/circl/kem/mlkem/mlkem768"
	"golang.org/x/crypto/sha3"
)

var kemScheme = mlkem768.Scheme()

// Note represents an unspent confidential value record in the OENEXA shielded pool.
type Note struct {
	Value       uint64
	RecipientPK []byte   // recipient's public key (ML-DSA-65 or address)
	Rho         [32]byte // unique random serial number used to compute nullifier
	Rcm         [32]byte // random commitment trapdoor
	Memo        []byte   // optional encrypted memo text
}

// NewNote creates a fresh shielded note with cryptographically secure random Rho and Rcm.
func NewNote(recipientPK []byte, value uint64, memo []byte) (*Note, error) {
	if len(recipientPK) == 0 {
		return nil, errors.New("shielded: empty recipient public key")
	}
	var rho, rcm [32]byte
	if _, err := io.ReadFull(rand.Reader, rho[:]); err != nil {
		return nil, fmt.Errorf("shielded: rand rho: %w", err)
	}
	if _, err := io.ReadFull(rand.Reader, rcm[:]); err != nil {
		return nil, fmt.Errorf("shielded: rand rcm: %w", err)
	}
	return &Note{
		Value:       value,
		RecipientPK: recipientPK,
		Rho:         rho,
		Rcm:         rcm,
		Memo:        memo,
	}, nil
}

// Commitment calculates the quantum-safe SHA-3-256 note commitment:
// cm = SHA-3-256("OENEXA_NOTE_CM" || value || RecipientPK || Rho || Rcm)
func (n *Note) Commitment() [32]byte {
	h := sha3.New256()
	h.Write([]byte("OENEXA_NOTE_CM"))
	var b8 [8]byte
	binary.BigEndian.PutUint64(b8[:], n.Value)
	h.Write(b8[:])
	h.Write(n.RecipientPK)
	h.Write(n.Rho[:])
	h.Write(n.Rcm[:])
	var cm [32]byte
	copy(cm[:], h.Sum(nil))
	return cm
}

// Serialize encodes the note fields into canonical bytes.
func (n *Note) Serialize() []byte {
	var buf bytes.Buffer
	var b8 [8]byte
	binary.BigEndian.PutUint64(b8[:], n.Value)
	buf.Write(b8[:])
	buf.Write(n.Rho[:])
	buf.Write(n.Rcm[:])
	var b4 [4]byte
	binary.BigEndian.PutUint32(b4[:], uint32(len(n.RecipientPK)))
	buf.Write(b4[:])
	buf.Write(n.RecipientPK)
	binary.BigEndian.PutUint32(b4[:], uint32(len(n.Memo)))
	buf.Write(b4[:])
	buf.Write(n.Memo)
	return buf.Bytes()
}

var randReader io.Reader = rand.Reader

// DeserializeNote decodes canonical bytes into a Note.
func DeserializeNote(data []byte) (*Note, error) {
	const minHeader = 8 + 32 + 32 + 4 + 4 // val(8) + rho(32) + rcm(32) + pkLen(4) + memoLen(4)
	if len(data) < minHeader {
		return nil, errors.New("shielded: note payload too short")
	}
	val := binary.BigEndian.Uint64(data[0:8])
	var rho, rcm [32]byte
	copy(rho[:], data[8:40])
	copy(rcm[:], data[40:72])

	pkLen := binary.BigEndian.Uint32(data[72:76])
	offset := 76
	if uint64(len(data)-offset) < uint64(pkLen)+4 {
		return nil, errors.New("shielded: truncated recipient public key or memo length")
	}
	pk := make([]byte, pkLen)
	copy(pk, data[offset:offset+int(pkLen)])
	offset += int(pkLen)

	memoLen := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	if uint64(len(data)-offset) < uint64(memoLen) {
		return nil, errors.New("shielded: truncated memo payload")
	}
	memo := make([]byte, memoLen)
	copy(memo, data[offset:offset+int(memoLen)])

	return &Note{
		Value:       val,
		RecipientPK: pk,
		Rho:         rho,
		Rcm:         rcm,
		Memo:        memo,
	}, nil
}

// Encrypt encrypts the note for the recipient's ML-KEM-768 public key.
// Wire format: [4-byte big-endian CT len][KEM CT][12-byte GCM Nonce][GCM Ciphertext]
func (n *Note) Encrypt(recipientPubKEMBytes []byte) ([]byte, error) {
	pubKEM, err := kemScheme.UnmarshalBinaryPublicKey(recipientPubKEMBytes)
	if err != nil {
		return nil, fmt.Errorf("shielded: unmarshal KEM public key: %w", err)
	}
	ct, ss, _ := kemScheme.Encapsulate(pubKEM)

	block, _ := aes.NewCipher(ss[:32])
	gcm, _ := cipher.NewGCM(block)

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(randReader, nonce); err != nil {
		return nil, fmt.Errorf("shielded: rand nonce: %w", err)
	}

	plaintext := n.Serialize()
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	var out bytes.Buffer
	var b4 [4]byte
	binary.BigEndian.PutUint32(b4[:], uint32(len(ct)))
	out.Write(b4[:])
	out.Write(ct)
	out.Write(nonce)
	out.Write(ciphertext)
	return out.Bytes(), nil
}

// Decrypt decrypts an encrypted note using the recipient's ML-KEM-768 private key.
func Decrypt(encData []byte, recipientPrivKEMBytes []byte) (*Note, error) {
	if len(encData) < 4 {
		return nil, errors.New("shielded: encrypted data too short")
	}
	ctLen := int(binary.BigEndian.Uint32(encData[:4]))
	if len(encData) < 4+ctLen+12 {
		return nil, errors.New("shielded: corrupted encrypted payload")
	}

	ct := encData[4 : 4+ctLen]
	rest := encData[4+ctLen:]
	nonce := rest[:12]
	ciphertext := rest[12:]

	privKEM, err := kemScheme.UnmarshalBinaryPrivateKey(recipientPrivKEMBytes)
	if err != nil {
		return nil, fmt.Errorf("shielded: unmarshal KEM private key: %w", err)
	}
	ss, err := kemScheme.Decapsulate(privKEM, ct)
	if err != nil {
		return nil, fmt.Errorf("shielded: KEM decapsulate: %w", err)
	}

	block, _ := aes.NewCipher(ss[:32])
	gcm, _ := cipher.NewGCM(block)

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("shielded: decrypt gcm: %w", err)
	}
	return DeserializeNote(plaintext)
}
