package hash

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
)

func HashPass(pass string) (string, error) {
	key := []byte(os.Getenv("CRYPT_KEY"))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nil, nonce, []byte(pass), nil)
	result := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}
