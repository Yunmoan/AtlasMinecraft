package proto

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// Minecraft uses AES/CFB8/NoPadding with the shared secret as both key and IV.
// Go's cipher.NewCFBEncrypter is CFB128, so it cannot be used here.
type cfb8 struct {
	block   cipher.Block
	shift   [aes.BlockSize]byte
	decrypt bool
}

func newCFB8(block cipher.Block, secret []byte, decrypt bool) *cfb8 {
	s := &cfb8{block: block, decrypt: decrypt}
	copy(s.shift[:], secret)
	return s
}

func (s *cfb8) XORKeyStream(dst, src []byte) {
	if len(dst) < len(src) {
		panic("cfb8: output smaller than input")
	}
	var encrypted [aes.BlockSize]byte
	for i, input := range src {
		s.block.Encrypt(encrypted[:], s.shift[:])
		output := input ^ encrypted[0]
		ciphertext := output
		if s.decrypt {
			ciphertext = input
		}
		copy(s.shift[:], s.shift[1:])
		s.shift[aes.BlockSize-1] = ciphertext
		dst[i] = output
	}
}

// EnableEncryption switches both packet directions after the unencrypted
// Encryption Response. Reading through the old buffered reader preserves any
// encrypted bytes it may already have read ahead from the TCP connection.
func (f *FramedConn) EnableEncryption(secret []byte) error {
	if len(secret) != aes.BlockSize {
		return errors.New("Minecraft shared secret must be 16 bytes")
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		return err
	}
	f.rd = bufio.NewReaderSize(&cipher.StreamReader{S: newCFB8(block, secret, true), R: f.rd}, 1<<16)
	f.wr = &cipher.StreamWriter{S: newCFB8(block, secret, false), W: f.Conn}
	return nil
}
