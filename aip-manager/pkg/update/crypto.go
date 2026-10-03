package update

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	SaltSize      = 16
	IVSize        = 12
	KDFIterations = 480000
	AESKeyLength  = 32
	HeaderSize    = 8
)

func PBKDF2SHA256(password []byte, salt []byte, iter int, keyLen int) []byte {
	h := hmac.New(sha256.New, password)
	hashSize := sha256.Size
	numBlocks := (keyLen + hashSize - 1) / hashSize
	var result []byte

	for block := 1; block <= numBlocks; block++ {
		h.Reset()
		h.Write(salt)
		var blockBuf [4]byte
		binary.BigEndian.PutUint32(blockBuf[:], uint32(block))
		h.Write(blockBuf[:])
		u := h.Sum(nil)

		blockSum := make([]byte, len(u))
		copy(blockSum, u)

		for i := 2; i <= iter; i++ {
			h.Reset()
			h.Write(u)
			u = h.Sum(nil)
			for k := 0; k < len(blockSum); k++ {
				blockSum[k] ^= u[k]
			}
		}
		result = append(result, blockSum...)
	}
	return result[:keyLen]
}

func EncryptFile(plaintext []byte, password string) ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	iv := make([]byte, IVSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	key := PBKDF2SHA256([]byte(password), salt, KDFIterations, AESKeyLength)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	ciphertext := aesgcm.Seal(nil, iv, plaintext, nil)

	var output []byte
	output = append(output, salt...)
	output = append(output, iv...)
	output = append(output, ciphertext...)

	return output, nil
}

func SignData(encryptedData []byte, privKey *rsa.PrivateKey) ([]byte, error) {
	hash := sha256.Sum256(encryptedData)

	opts := &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthAuto,
		Hash:       crypto.SHA256,
	}

	signature, err := rsa.SignPSS(rand.Reader, privKey, hash[:], opts)
	if err != nil {
		return nil, fmt.Errorf("failed to sign digest: %w", err)
	}

	return signature, nil
}

func ParsePrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := keyInterface.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, errors.New("key in PKCS#8 block is not an RSA private key")
	}

	return nil, fmt.Errorf("unsupported or encrypted private key format: %w", err)
}

func BuildPackage(inputFile, keyPath, password, outputPath string) error {
	rawBytes, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	privKey, err := ParsePrivateKey(keyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	encryptedData, err := EncryptFile(rawBytes, password)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	signature, err := SignData(encryptedData, privKey)
	if err != nil {
		return fmt.Errorf("signing failed: %w", err)
	}

	sigLen := uint64(len(signature))
	pkgFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output package: %w", err)
	}
	defer pkgFile.Close()

	if err := binary.Write(pkgFile, binary.BigEndian, sigLen); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := pkgFile.Write(signature); err != nil {
		return fmt.Errorf("failed to write signature: %w", err)
	}

	if _, err := pkgFile.Write(encryptedData); err != nil {
		return fmt.Errorf("failed to write encrypted data: %w", err)
	}

	return nil
}