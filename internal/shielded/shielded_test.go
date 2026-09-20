package shielded

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"io"
	"testing"
)

func TestNote_Lifecycle(t *testing.T) {
	// 1. Empty recipient PK error
	_, err := NewNote(nil, 100, []byte("memo"))
	if err == nil {
		t.Fatal("expected error for empty recipient PK")
	}

	// 2. Valid creation
	pk := []byte("dummy-ml-dsa-pk")
	note, err := NewNote(pk, 5000, []byte("invoice-101"))
	if err != nil {
		t.Fatalf("NewNote failed: %v", err)
	}
	if note.Value != 5000 {
		t.Errorf("expected value 5000, got %d", note.Value)
	}

	// 3. Commitment deterministic
	cm1 := note.Commitment()
	cm2 := note.Commitment()
	if cm1 != cm2 {
		t.Fatal("note commitment is not deterministic")
	}

	// 4. Serialization roundtrip
	data := note.Serialize()
	deserialized, err := DeserializeNote(data)
	if err != nil {
		t.Fatalf("DeserializeNote failed: %v", err)
	}
	if deserialized.Value != note.Value || deserialized.Commitment() != note.Commitment() {
		t.Fatal("deserialized note mismatch")
	}

	// 5. Corrupt deserialization
	_, err = DeserializeNote([]byte("too-short"))
	if err == nil {
		t.Fatal("expected error on short note bytes")
	}
	// Truncated at various points
	_, err = DeserializeNote(data[:8+32+32+2])
	if err == nil {
		t.Fatal("expected error on truncated note bytes")
	}

	// Truncated recipient PK
	var truncatedBuf [80]byte
	binary.BigEndian.PutUint32(truncatedBuf[72:76], 10)
	_, err = DeserializeNote(truncatedBuf[:])
	if err == nil {
		t.Fatal("expected error on truncated recipient PK")
	}

	// Truncated memo
	var truncatedMemoBuf [80]byte
	binary.BigEndian.PutUint32(truncatedMemoBuf[72:76], 0)
	binary.BigEndian.PutUint32(truncatedMemoBuf[76:80], 10)
	_, err = DeserializeNote(truncatedMemoBuf[:])
	if err == nil {
		t.Fatal("expected error on truncated memo")
	}
}

func TestNote_EncryptionRoundtrip(t *testing.T) {
	pub, priv, err := kemScheme.GenerateKeyPair()
	if err != nil {
		t.Fatalf("kem GenerateKeyPair: %v", err)
	}
	pubBytes, _ := pub.MarshalBinary()
	privBytes, _ := priv.MarshalBinary()

	note, _ := NewNote([]byte("recipient-pk"), 10_000, []byte("confidential"))

	// 1. Valid Encrypt and Decrypt
	enc, err := note.Encrypt(pubBytes)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decNote, err := Decrypt(enc, privBytes)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decNote.Value != note.Value || decNote.Commitment() != note.Commitment() {
		t.Fatal("decrypted note does not match original")
	}

	// 2. Encrypt with invalid KEM pubkey
	_, err = note.Encrypt([]byte("invalid-kem-pubkey"))
	if err == nil {
		t.Fatal("expected error encrypting with invalid pubkey")
	}

	// 2b. Encrypt with failing randReader for nonce
	origRandReader := randReader
	randReader = &mockFailingReader{failAfter: 0}
	_, err = note.Encrypt(pubBytes)
	randReader = origRandReader
	if err == nil {
		t.Fatal("expected error when randReader fails in Encrypt")
	}

	// 3. Decrypt with invalid short data
	_, err = Decrypt([]byte{0, 1}, privBytes)
	if err == nil {
		t.Fatal("expected error decrypting short data")
	}

	// 4. Decrypt with corrupted length
	badLen := []byte{0xFF, 0xFF, 0, 0, 1, 2, 3}
	_, err = Decrypt(badLen, privBytes)
	if err == nil {
		t.Fatal("expected error decrypting corrupted length")
	}

	// 5. Decrypt with invalid KEM privkey
	_, err = Decrypt(enc, []byte("bad-priv-key"))
	if err == nil {
		t.Fatal("expected error decrypting with bad private key")
	}

	// 5b. Decrypt with invalid ciphertext length that causes KEM decapsulate error
	badCT := make([]byte, 4+10+12+16)
	binary.BigEndian.PutUint32(badCT[:4], 10)
	_, err = Decrypt(badCT, privBytes)
	if err == nil {
		t.Fatal("expected error when KEM decapsulate fails")
	}

	// 6. Decrypt with wrong key (GCM authentication failure)
	_, otherPriv, _ := kemScheme.GenerateKeyPair()
	otherPrivBytes, _ := otherPriv.MarshalBinary()
	_, err = Decrypt(enc, otherPrivBytes)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}

	// 7. Decrypt corrupted ciphertext bytes (GCM authentication failure)
	corruptEnc := make([]byte, len(enc))
	copy(corruptEnc, enc)
	corruptEnc[len(corruptEnc)-1] ^= 0xFF
	_, err = Decrypt(corruptEnc, privBytes)
	if err == nil {
		t.Fatal("expected error on corrupted ciphertext bytes")
	}

	// 8. Decrypt payload where decrypted plaintext is invalid Note
	pubKEM, _ := kemScheme.UnmarshalBinaryPublicKey(pubBytes)
	ct, ss, _ := kemScheme.Encapsulate(pubKEM)
	block, _ := aes.NewCipher(ss[:32])
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, 12)
	badPlaintextCiphertext := gcm.Seal(nil, nonce, []byte("too short"), nil)
	var badPlaintextPayload bytes.Buffer
	var b4 [4]byte
	binary.BigEndian.PutUint32(b4[:], uint32(len(ct)))
	badPlaintextPayload.Write(b4[:])
	badPlaintextPayload.Write(ct)
	badPlaintextPayload.Write(nonce)
	badPlaintextPayload.Write(badPlaintextCiphertext)
	_, err = Decrypt(badPlaintextPayload.Bytes(), privBytes)
	if err == nil {
		t.Fatal("expected error when decrypted plaintext cannot be deserialized as a note")
	}
}

func TestNullifier_Operations(t *testing.T) {
	sk := []byte("secret-key")
	var rho, cm [32]byte
	rho[0] = 1
	cm[0] = 2

	nf1 := ComputeNullifier(sk, rho, cm)
	nf2 := ComputeNullifier(sk, rho, cm)
	if nf1 != nf2 {
		t.Fatal("nullifier not deterministic")
	}

	ns := NewNullifierSet()
	if ns.Has(nf1) {
		t.Fatal("unspent nullifier marked as spent")
	}
	if err := ns.Add(nf1); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if !ns.Has(nf1) {
		t.Fatal("spent nullifier not found")
	}
	if ns.Len() != 1 {
		t.Errorf("expected count 1, got %d", ns.Len())
	}

	// Double-spend rejection
	err := ns.Add(nf1)
	if err == nil {
		t.Fatal("expected error on double-spend Add")
	}

	// Clone
	c := ns.Clone()
	if !c.Has(nf1) || c.Len() != 1 {
		t.Fatal("clone mismatch")
	}
}

func TestTree_Accumulation(t *testing.T) {
	tree := NewNoteCommitmentTree()
	if tree.LeafCount() != 0 {
		t.Fatal("new tree should have 0 leaves")
	}
	emptyRoot := tree.Root()

	var leaves [5][32]byte
	for i := 0; i < 5; i++ {
		leaves[i][0] = byte(i + 1)
		idx, err := tree.Append(leaves[i])
		if err != nil {
			t.Fatalf("Append %d failed: %v", i, err)
		}
		if idx != uint32(i) {
			t.Errorf("expected idx %d, got %d", i, idx)
		}
	}
	if tree.LeafCount() != 5 {
		t.Errorf("expected 5 leaves, got %d", tree.LeafCount())
	}

	root := tree.Root()
	if root == emptyRoot {
		t.Fatal("root should have changed after appends")
	}

	// Witness generation and verification for all leaves
	for i := uint32(0); i < 5; i++ {
		path, err := tree.Witness(i)
		if err != nil {
			t.Fatalf("Witness %d failed: %v", i, err)
		}
		if !VerifyWitness(leaves[i], i, root, path) {
			t.Fatalf("VerifyWitness failed for leaf %d", i)
		}
		// Sibling verification with wrong leaf should fail
		var wrongLeaf [32]byte
		wrongLeaf[0] = 99
		if VerifyWitness(wrongLeaf, i, root, path) {
			t.Fatalf("VerifyWitness should fail with wrong leaf for index %d", i)
		}
	}

	// Witness out of bounds
	_, err := tree.Witness(10)
	if err == nil {
		t.Fatal("expected error on out of bounds witness")
	}

	// VerifyWitness with invalid path length
	if VerifyWitness(leaves[0], 0, root, nil) {
		t.Fatal("VerifyWitness should return false for invalid path length")
	}

	// Clone check
	treeClone := tree.Clone()
	if treeClone.LeafCount() != 5 || treeClone.Root() != root {
		t.Fatal("tree clone mismatch")
	}
}

func TestTurnstile_SupplyConservation(t *testing.T) {
	total := uint64(1_000_000)
	ts := NewTurnstile(total)

	// Invariant hold initially
	if err := ts.VerifyInvariant(total); err != nil {
		t.Fatalf("initial invariant check failed: %v", err)
	}

	// Shield 300,000
	if err := ts.Shield(300_000); err != nil {
		t.Fatalf("Shield failed: %v", err)
	}
	tr, sh := ts.Balances()
	if tr != 700_000 || sh != 300_000 {
		t.Errorf("unexpected balances: tr=%d, sh=%d", tr, sh)
	}
	if err := ts.VerifyInvariant(total); err != nil {
		t.Fatalf("post-shield invariant failed: %v", err)
	}

	// Shield zero / excessive error
	if err := ts.Shield(0); err == nil {
		t.Fatal("expected error on shield 0")
	}
	if err := ts.Shield(800_000); err == nil {
		t.Fatal("expected error on excessive shield")
	}

	// Unshield 100,000
	if err := ts.Unshield(100_000); err != nil {
		t.Fatalf("Unshield failed: %v", err)
	}
	tr, sh = ts.Balances()
	if tr != 800_000 || sh != 200_000 {
		t.Errorf("unexpected balances after unshield: tr=%d, sh=%d", tr, sh)
	}
	if err := ts.VerifyInvariant(total); err != nil {
		t.Fatalf("post-unshield invariant failed: %v", err)
	}

	// Unshield zero / excessive error
	if err := ts.Unshield(0); err == nil {
		t.Fatal("expected error on unshield 0")
	}
	if err := ts.Unshield(500_000); err == nil {
		t.Fatal("expected error on excessive unshield")
	}

	// Invariant violation check
	if err := ts.VerifyInvariant(999_999); err == nil {
		t.Fatal("expected error on invariant violation")
	}

	// Snapshot
	snap := ts.Snapshot()
	str, ssh := snap.Balances()
	if str != tr || ssh != sh {
		t.Fatal("snapshot mismatch")
	}
}

func TestViewingKey_PaymentDisclosure(t *testing.T) {
	sk := []byte("spending-key-1234")
	vk := DeriveViewingKey(sk)

	note, _ := NewNote([]byte("merchant-pk"), 1500, []byte("order-888"))
	var txHash [32]byte
	txHash[0] = 0xAA

	// 1. Create and verify payment proof
	proof, err := CreatePaymentProof(note, sk, txHash)
	if err != nil {
		t.Fatalf("CreatePaymentProof failed: %v", err)
	}
	if !VerifyPaymentProof(proof, vk, txHash, []byte("merchant-pk"), 1500) {
		t.Fatal("valid payment proof failed verification")
	}

	// 2. Error on nil note or empty spending key
	_, err = CreatePaymentProof(nil, sk, txHash)
	if err == nil {
		t.Fatal("expected error for nil note")
	}
	_, err = CreatePaymentProof(note, nil, txHash)
	if err == nil {
		t.Fatal("expected error for empty spending key")
	}

	// 3. Verify failure cases
	if VerifyPaymentProof(nil, vk, txHash, []byte("merchant-pk"), 1500) {
		t.Fatal("expected false for nil proof")
	}
	var wrongTxHash [32]byte
	if VerifyPaymentProof(proof, vk, wrongTxHash, []byte("merchant-pk"), 1500) {
		t.Fatal("expected false for wrong tx hash")
	}
	if VerifyPaymentProof(proof, vk, txHash, []byte("merchant-pk"), 9999) {
		t.Fatal("expected false for wrong value")
	}
	if VerifyPaymentProof(proof, vk, txHash, []byte("wrong-recipient"), 1500) {
		t.Fatal("expected false for wrong recipient PK")
	}
	otherVK := DeriveViewingKey([]byte("other-key"))
	if VerifyPaymentProof(proof, otherVK, txHash, []byte("merchant-pk"), 1500) {
		t.Fatal("expected false for wrong viewing key")
	}
}

// Edge case for full tree error check
func TestTree_FullError(t *testing.T) {
	origMax := MaxTreeLeaves
	MaxTreeLeaves = 1
	defer func() { MaxTreeLeaves = origMax }()

	tree := NewNoteCommitmentTree()
	var cm [32]byte
	cm[0] = 1
	if _, err := tree.Append(cm); err != nil {
		t.Fatalf("first append failed: %v", err)
	}
	if _, err := tree.Append(cm); err == nil {
		t.Fatal("expected error on full tree append")
	}
}

func TestNote_RandErrors(t *testing.T) {
	// Exercise NewNote error branches using mock reader
	origReader := rand.Reader
	defer func() { rand.Reader = origReader }()

	rand.Reader = &mockFailingReader{failAfter: 0}
	_, err := NewNote([]byte("pk"), 100, nil)
	if err == nil {
		t.Fatal("expected error on failing rand reader for rho")
	}

	rand.Reader = &mockFailingReader{failAfter: 32}
	_, err = NewNote([]byte("pk"), 100, nil)
	if err == nil {
		t.Fatal("expected error on failing rand reader for rcm")
	}
}

type mockFailingReader struct {
	readBytes int
	failAfter int
}

func (m *mockFailingReader) Read(p []byte) (n int, err error) {
	if m.readBytes >= m.failAfter {
		return 0, io.ErrUnexpectedEOF
	}
	toRead := len(p)
	if m.readBytes+toRead > m.failAfter {
		toRead = m.failAfter - m.readBytes
	}
	for i := 0; i < toRead; i++ {
		p[i] = 1
	}
	m.readBytes += toRead
	if m.readBytes >= m.failAfter && toRead < len(p) {
		return toRead, io.ErrUnexpectedEOF
	}
	return toRead, nil
}
