package util

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	_ "crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
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
	iv, e := GenerateRandomBytes(aes.BlockSize)

	mode := cipher.NewCBCEncrypter(block, iv) //, iv []byte) NewGCM(block) //, iv)

	pt, e := Pkcs7Pad([]byte(plaintext), aes.BlockSize)
	ct := make([]byte, len(pt))

	mode.CryptBlocks(ct, pt)

	ct = append(iv, ct...)

	//authenticate
	mac := hmac.New(crypto.SHA256.New, key)
	mac.Write(ct)
	mmac := mac.Sum(nil)

	ct = append(mmac, ct...)

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
	//check MAC before anything else
	ctmac := ct[:crypto.SHA256.Size()]

	mac := hmac.New(crypto.SHA256.New, key)
	mac.Write(ct[crypto.SHA256.Size():])
	ourmac := mac.Sum(nil)

	if !hmac.Equal(ctmac, ourmac) {
		return []byte{}, errors.New("HMAC Fail")
	}

	ct = ct[crypto.SHA256.Size():]

	block, e := aes.NewCipher(key)
	if e != nil {
		return []byte{}, e
	}

	mode := cipher.NewCBCDecrypter(block, ct[:aes.BlockSize]) //NewGCM(block) //, iv)
	if e != nil {
		return []byte{}, e
	}

	ct = ct[aes.BlockSize:]
	pt := make([]byte, len(ct))

	mode.CryptBlocks(pt, ct)

	//unpad
	pt, e = Pkcs7Unpad(pt, aes.BlockSize)

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

//https://github.com/go-web/tokenizer/blob/master/pkcs7.go

var (
	// ErrInvalidBlockSize indicates hash blocksize <= 0.
	ErrInvalidBlockSize = errors.New("invalid blocksize")

	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")

	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
)

func Pkcs7Pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

func Pkcs7Unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	if len(b)%blocksize != 0 {
		return nil, ErrInvalidPKCS7Padding
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		return nil, ErrInvalidPKCS7Padding
	}
	for i := 0; i < n; i++ {
		if b[len(b)-n+i] != c {
			return nil, ErrInvalidPKCS7Padding
		}
	}
	return b[:len(b)-n], nil
}
