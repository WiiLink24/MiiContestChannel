package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"github.com/wii-tools/lzx/lz10"
	"os"
	"path/filepath"
)

var (
	aesKey  = []byte{0x8D, 0x22, 0xA3, 0xD8, 0x08, 0xD5, 0xD0, 0x72, 0x02, 0x74, 0x36, 0xB6, 0x30, 0x3C, 0x5B, 0x50}
	aesIv   = []byte{0xBE, 0x5E, 0x54, 0x89, 0x25, 0xAC, 0xDD, 0x3C, 0xD5, 0x34, 0x2E, 0x08, 0xFB, 0x8A, 0xBF, 0xEC}
	hmacKey = []byte{0x4C, 0xC0, 0x8F, 0xA1, 0x41, 0xDE, 0x25, 0x37, 0xAA, 0xA5, 0x2B, 0x8D, 0xAC, 0xD9, 0xB5, 0x63, 0x35, 0xAF, 0xE4, 0x67}
)

// writerState helps keep track of internal state of the writing process.
type writerState struct {
	compressedData []byte
	encryptedData  []byte
	hmacSignature  []byte
}

// Write packs inputted data into a format that the Mii Contest Channel wants.
// This includes an "MC" header, HMAC-SHA1 signature, as well as compressed and encrypted data.
func Write(data Writable, path string) error {
	var err error
	writer := new(writerState)

	writer.compressedData, err = lz10.Compress(data.ToBytes(data))
	if err != nil {
		return err
	}

	for len(writer.compressedData)%aes.BlockSize != 0 {
		writer.compressedData = append(writer.compressedData, 0)
	}

	err = writer.Encrypt()
	if err != nil {
		return err
	}

	writer.MakeHMACSignature()

	err = writer.Write(path)
	if err != nil {
		return err
	}

	return nil
}

// Encrypt encrypts the compressed data in AES-128-CBC.
func (w *writerState) Encrypt() error {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return err
	}

	w.encryptedData = make([]byte, len(w.compressedData))
	mode := cipher.NewCBCEncrypter(block, aesIv)
	mode.CryptBlocks(w.encryptedData, w.compressedData)

	return nil
}

// MakeHMACSignature takes the encrypted data and calculates the HMAC-SHA1 signature.
func (w *writerState) MakeHMACSignature() {
	h := hmac.New(sha1.New, hmacKey)
	h.Write(w.encryptedData)
	w.hmacSignature = h.Sum(nil)
}

// Write does all the writing of the data in writerState.
func (w *writerState) Write(path string) error {
	buffer := new(bytes.Buffer)

	// Header
	buffer.WriteString("MC")
	buffer.Write([]byte{0, 1})

	// HMAC-SHA1
	buffer.Write(w.hmacSignature)

	// Encrypted data
	buffer.Write(w.encryptedData)

	// Create directories if they don't exist
	filePath := fmt.Sprintf("%s/%s", GetConfig().AssetsPath, path)
	err := os.MkdirAll(filepath.Dir(filePath), 0775)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, buffer.Bytes(), 0664)
}
