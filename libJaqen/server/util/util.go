package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

//https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

//EncryptString will return an AES encrypted byte array of the given plaintext string, using the key and IV supplied
//the Ciphertext will have the IV/Nonce prefixed to the blob
func EncryptString(plaintext string, key []byte) ([]byte, error) {
	block, e := aes.NewCipher(key)
	if e != nil {
		return []byte{}, e
	}

	mode, e := cipher.NewGCM(block) //, iv)

	iv, e := GenerateRandomBytes(mode.NonceSize())

	//mode.CryptBlocks(ct, []byte(plaintext))
	ct := mode.Seal(nil, iv, []byte(plaintext), nil)

	ct = append(iv, ct...)
	return ct, nil
}

//EncryptStringToHex will take a plaintext, key and IV, and encrypt (using CBC), returning a hex encoded string
func EncryptStringToHex(plaintext string, key []byte) (string, error) {
	b, e := EncryptString(plaintext, key)
	if e != nil {
		return "", e
	}
	s := hex.EncodeToString(b)
	return s, e
}

func Decrypt(ct, key []byte) ([]byte, error) {
	block, e := aes.NewCipher(key)
	if e != nil {
		return []byte{}, e
	}
	mode, e := cipher.NewGCM(block) //, iv)
	if e != nil {
		return []byte{}, e
	}
	iv := ct[:mode.NonceSize()]
	ct = ct[mode.NonceSize():]

	pt := []byte{}

	pt, e = mode.Open(nil, iv, ct, nil)
	//mode.CryptBlocks(pt, ct)

	return pt, e
}

func DecryptToString(ct, key []byte) (string, error) {
	pt, e := Decrypt(ct, key)
	if e != nil {
		return "", e
	}
	return string(pt), nil
}

func DecryptHexStringToString(ct string, key []byte) (string, error) {
	decodedCt, e := hex.DecodeString(ct)
	if e != nil {
		return "", e
	}

	return DecryptToString(decodedCt, key)
}
