// pkg/crypto/encrypt.go
package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
)

// --- Cryptographic Constants ---
const (
	// (16 bytes for AES-128)
	aesKeyString = "G7k!qZ2@vP9x#L4m"
)

// IV (Initialization Vector) must be 16 bytes for AES.
// The C++ code `const byte iv[AES::BLOCKSIZE] = "111111111111111";`
// initializes a 16-byte array with the 15 '1' characters followed by a null terminator ('\0').
// This is the exact equivalent in Go.
var iv = []byte{
	'1', '1', '1', '1', '1', '1', '1', '1',
	'1', '1', '1', '1', '1', '1', '1', 0,
}

// pkcs7Pad pads data to a multiple of blockSize using PKCS#7 padding.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpad removes PKCS#7 padding from the data.
func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: unpadding error, input is empty")
	}
	unpadding := int(data[length-1])
	if unpadding > length || unpadding == 0 {
		return nil, fmt.Errorf("pkcs7: unpadding error, invalid padding size (%d)", unpadding)
	}
	for i := 0; i < unpadding; i++ {
		if data[length-unpadding+i] != byte(unpadding) {
			return nil, errors.New("pkcs7: unpadding error, invalid padding bytes")
		}
	}
	return data[:(length - unpadding)], nil
}

// Encrypt performs AES-128-CBC encryption with PKCS#7 padding.
func Encrypt(plainText string) (string, error) {
	key := []byte(aesKeyString)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	paddedPlaintext := pkcs7Pad([]byte(plainText), aes.BlockSize)
	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	hexString := hex.EncodeToString(ciphertext)
	return "#x" + hexString, nil
}

// Decrypt performs AES-128-CBC decryption with PKCS#7 padding.
func Decrypt(hexCiphertext string, key string) (string, error) {
	// Remove the "#x" prefix if it exists
	if len(hexCiphertext) > 2 && hexCiphertext[:2] == "#x" {
		hexCiphertext = hexCiphertext[2:]
	}

	ciphertext, err := hex.DecodeString(hexCiphertext)
	if err != nil {
		return "", err
	}

	keyBytes := []byte(key)
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	recoveredBytes := make([]byte, len(ciphertext))
	mode.CryptBlocks(recoveredBytes, ciphertext)

	unpaddedBytes, err := pkcs7Unpad(recoveredBytes)
	if err != nil {
		return "", err
	}

	return string(unpaddedBytes), nil
}
