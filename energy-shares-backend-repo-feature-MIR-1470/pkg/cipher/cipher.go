package cipher

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"
)

const (
	_CipherKey   = "wM3b*MLjg9LHLrfXWbC-iR3hbjxz!3n8" //32文字のみ可能
	_CipherIvKey = "AHPQ@a!N7xoBE_Ur"                 //16文字のみ可能
)

type Crypto interface {
	Encrypt(plainText string) (string, error)
	Decrypt(cipherIvKey string) (string, error)
}

type niceCrypto struct {
	cipherKey   string
	cipherIvKey string
}

func (c niceCrypto) Encrypt(plainText string) (string, error) {
	if strings.TrimSpace(plainText) == "" {
		return plainText, nil
	}

	block, err := aes.NewCipher([]byte(c.cipherKey))
	if err != nil {
		return "", err
	}

	encrypter := cipher.NewCBCEncrypter(block, []byte(c.cipherIvKey))
	paddedPlainText := padPKCS7([]byte(plainText), encrypter.BlockSize())

	cipherText := make([]byte, len(paddedPlainText))

	encrypter.CryptBlocks(cipherText, paddedPlainText)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func (c niceCrypto) Decrypt(cipherText string) (string, error) {
	if strings.TrimSpace(cipherText) == "" {
		return cipherText, nil
	}

	decodedCipherText, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(c.cipherKey))
	if err != nil {
		return "", err
	}

	decrypter := cipher.NewCBCDecrypter(block, []byte(c.cipherIvKey))
	plainText := make([]byte, len(decodedCipherText))

	decrypter.CryptBlocks(plainText, decodedCipherText)
	trimmedPlainText := trimPKCS5(plainText)

	return string(trimmedPlainText), nil
}

// NewNiceCrypto deprecated
func NewNiceCrypto(cipherKey, cipherIvKey string) (Crypto, error) {
	if ck := len(cipherKey); ck != 32 {
		return nil, aes.KeySizeError(ck)
	}

	if cik := len(cipherIvKey); cik != 16 {
		return nil, aes.KeySizeError(cik)
	}

	return &niceCrypto{cipherKey, cipherIvKey}, nil
}

func New() (Crypto, error) {
	if ck := len(_CipherKey); ck != 32 {
		return nil, aes.KeySizeError(ck)
	}

	if cik := len(_CipherIvKey); cik != 16 {
		return nil, aes.KeySizeError(cik)
	}

	return &niceCrypto{_CipherKey, _CipherIvKey}, nil
}

func padPKCS7(plainText []byte, blockSize int) []byte {
	padding := blockSize - len(plainText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plainText, padText...)
}

func trimPKCS5(text []byte) []byte {
	padding := text[len(text)-1]
	return text[:len(text)-int(padding)]
}
