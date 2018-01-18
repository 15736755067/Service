// Make encryption in Go easy
package simpleaes

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"sync"
)

type cbcMode interface {
	cipher.BlockMode
	SetIV([]byte)
}

type Aes struct {
	em       sync.Mutex
	dm       sync.Mutex
	iv       []byte
	enc, dec cbcMode
}

// Creates a new encryption/decryption object
// with a given key of a given size
// (16, 24 or 32 for AES-128, AES-192 and AES-256 respectively,
// as per http://golang.org/pkg/crypto/aes/#NewCipher)
//
// The key will be padded to the given size if needed.
// An IV is created as a series of NULL bytes of necessary length
// when there is no iv string passed as 3rd value to function.
func New(size int, key string, more ...string) (*Aes, error) {
	padded := make([]byte, size)
	copy(padded, []byte(key))
	var iv []byte
	if len(more) > 0 {
		iv = []byte(more[0])
	} else {
		iv = make([]byte, size)
	}
	aes, err := aes.NewCipher(padded)
	if err != nil {
		return nil, err
	}
	enc := cipher.NewCBCEncrypter(aes, iv).(cbcMode)
	dec := cipher.NewCBCDecrypter(aes, iv).(cbcMode)
	return &Aes{sync.Mutex{}, sync.Mutex{}, iv, enc, dec}, nil
}

func (me *Aes) padSlice(src []byte) []byte {
	// src must be a multiple of block size
	bs := me.enc.BlockSize()
	mult := int((len(src) / bs) + 1)
	leng := bs * mult

	src_padded := make([]byte, leng)
	copy(src_padded, src)
	return src_padded
}

// Encrypt a slice of bytes, producing a new, freshly allocated slice
//
// Source will be padded with null bytes if necessary
func (me *Aes) Encrypt(src []byte, next ...bool) []byte {
	if len(src)%me.enc.BlockSize() != 0 {
		src = me.padSlice(src)
	}
	dst := make([]byte, len(src))
	me.em.Lock()
	defer me.em.Unlock()
	if len(next) == 0 || !next[0] {
		me.enc.SetIV(me.iv)
	}
	me.enc.CryptBlocks(dst, src)
	return dst
}

// Encrypt blocks from reader, write results into writer
func (me *Aes) EncryptStream(reader io.Reader, writer io.Writer) error {
	for {
		buf := make([]byte, me.enc.BlockSize())
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			if err == io.EOF {
				break
			} else if err == io.ErrUnexpectedEOF {
				// nothing
			} else {
				return err
			}
		}
		me.em.Lock()
		defer me.em.Unlock()
		me.enc.CryptBlocks(buf, buf)
		if _, err = writer.Write(buf); err != nil {
			return err
		}
	}
	return nil
}

// Decrypt a slice of bytes, producing a new, freshly allocated slice
//
// Source will be padded with null bytes if necessary
func (me *Aes) Decrypt(src []byte, next ...bool) []byte {
	if len(src)%me.dec.BlockSize() != 0 {
		src = me.padSlice(src)
	}
	dst := make([]byte, len(src))
	me.dm.Lock()
	defer me.dm.Unlock()
	if len(next) == 0 || !next[0] {
		me.dec.SetIV(me.iv)
	}
	me.dec.CryptBlocks(dst, src)
	return dst
}

// Decrypt blocks from reader, write results into writer
func (me *Aes) DecryptStream(reader io.Reader, writer io.Writer) error {
	buf := make([]byte, me.dec.BlockSize())
	for {
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		me.dm.Lock()
		defer me.dm.Unlock()
		me.dec.CryptBlocks(buf, buf)
		if _, err = writer.Write(buf); err != nil {
			return err
		}
	}
	return nil
}
