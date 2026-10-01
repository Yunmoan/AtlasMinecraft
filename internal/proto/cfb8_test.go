package proto

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"testing"
)

func TestMinecraftCFB8AgainstIndependentAESVector(t *testing.T) {
	secret := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	block, err := aes.NewCipher(secret)
	if err != nil {
		t.Fatal(err)
	}
	plain := make([]byte, 64)
	for i := range plain {
		plain[i] = byte(i)
	}
	want, _ := hex.DecodeString("0a22f796e1b93e9032cff804838adfc3a5e4b3ffdd47108575533e672ef5d8ef4929ef289ace5bca4e6d3ae9a7c1c02d6808e094e196bdaf5a6fe1f0f413cacd")
	got := make([]byte, len(plain))
	encrypt := newCFB8(block, secret, false)
	encrypt.XORKeyStream(got[:19], plain[:19])
	encrypt.XORKeyStream(got[19:], plain[19:])
	if !bytes.Equal(got, want) {
		t.Fatalf("CFB8 ciphertext = %x, want %x", got, want)
	}
	decrypted := make([]byte, len(got))
	decrypt := newCFB8(block, secret, true)
	decrypt.XORKeyStream(decrypted, got)
	if !bytes.Equal(decrypted, plain) {
		t.Fatalf("CFB8 decryption = %x, want %x", decrypted, plain)
	}
}
