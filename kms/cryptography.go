package kms

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/gob"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

type Key struct {
	PrivateKey string
	PublicKey  string
	Algorithm  string
}

func generateKey(password, algorithm string) (*Key, error) {
	key := &Key{}
	switch algorithm {
	case "secp256k1":
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		encryptedPrivKey, err := getPemEncodedPrivateKey(password, privateKey)
		if err != nil {
			return nil, err
		}
		encryptedPubKey, err := getPemEncodedPublicKey(password, privateKey.Public())
		if err != nil {
			return nil, err
		}
		key.Algorithm = algorithm
		key.PrivateKey = string(encryptedPrivKey)
		key.PublicKey = string(encryptedPubKey)
	case "ed25519":
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		encryptedPrivKey, err := getPemEncodedPrivateKey(password, privKey)
		if err != nil {
			return nil, err
		}
		encryptedPubKey, err := getPemEncodedPublicKey(password, pubKey)
		if err != nil {
			return nil, err
		}
		key.Algorithm = algorithm
		key.PrivateKey = string(encryptedPrivKey)
		key.PublicKey = string(encryptedPubKey)
	default:
		return nil, fmt.Errorf("invalid algorithm")
	}
	return key, nil
}

func getPemEncodedPrivateKey(password string, privKey interface{}) ([]byte, error) {
	bytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, err
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: bytes})
	return encryptData(pemEncoded, password)
}

func getPemEncodedPublicKey(password string, pubKey interface{}) ([]byte, error) {
	bytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, err
	}
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: bytes})
	return encryptData(pemEncodedPub, password)
}

func getDecodedPrivateKey(password string, privKey []byte) (interface{}, error) {
	pemEncoded, err := decryptData(privKey, password)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	return x509.ParsePKCS8PrivateKey(x509Encoded)
}

func getDecodedPublicKey(password string, publicKey []byte) (interface{}, error) {
	pemEncoded, err := decryptData(publicKey, password)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	return x509.ParsePKIXPublicKey(x509Encoded)
}

func encryptData(data []byte, keyPhrase string) ([]byte, error) {
	aesBlock, err := aes.NewCipher([]byte(mdHashing(keyPhrase)))
	if err != nil {
		return nil, err
	}

	gcmInstance, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcmInstance.NonceSize())
	cipheredText := gcmInstance.Seal(nonce, nonce, data, nil)

	return cipheredText, nil
}

func decryptData(cipherData []byte, keyPhrase string) ([]byte, error) {
	hashedPhrase := mdHashing(keyPhrase)
	aesBlock, err := aes.NewCipher([]byte(hashedPhrase))
	if err != nil {
		return nil, err
	}
	gcmInstance, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}
	nonceSize := gcmInstance.NonceSize()
	nonce, cipheredText := cipherData[:nonceSize], cipherData[nonceSize:]
	originalText, err := gcmInstance.Open(nil, nonce, cipheredText, nil)
	if err != nil {
		return nil, err
	}
	return originalText, nil
}

func mdHashing(input string) string {
	byteInput := []byte(input)
	md5Hash := md5.Sum(byteInput)
	return hex.EncodeToString(md5Hash[:])
}

func (k *Key) Hash() []byte {
	var b bytes.Buffer
	gob.NewEncoder(&b).Encode(k)
	return b.Bytes()
}
