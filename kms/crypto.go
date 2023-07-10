package kms

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"wallet-kms/vault"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type Wallet struct {
	Name      string
	Address   string
	Algorithm string
	WalletId  string
}

type ecPrivateKey struct {
	Version    int
	PrivateKey []byte
	PublicKey  asn1.BitString `asn1:"optional,explicit,tag:1"`
}

type pkcs8 struct {
	Version    int
	Algo       pkix.AlgorithmIdentifier
	PrivateKey []byte
	// optional attributes omitted.
}

type pkcs1PrivateKey struct {
	Version int
	N       *big.Int
	E       int
	D       *big.Int
	P       *big.Int
	Q       *big.Int
	// We ignore these values, if present, because rsa will calculate them.
	Dp   *big.Int `asn1:"optional"`
	Dq   *big.Int `asn1:"optional"`
	Qinv *big.Int `asn1:"optional"`

	AdditionalPrimes []pkcs1AdditionalRSAPrime `asn1:"optional,omitempty"`
}

type pkcs1AdditionalRSAPrime struct {
	Prime *big.Int

	// We ignore these values because rsa will calculate them.
	Exp   *big.Int
	Coeff *big.Int
}

const ecPrivKeyVersion = 1

func (w *Wallet) generateKey(ctx context.Context, vault vault.Vault) error {
	data, _ := vault.GetSecret(ctx, w.Name)
	if data != nil {
		return fmt.Errorf("key exist with specified name")
	}
	secret := make(map[string]interface{})
	switch w.Algorithm {
	case "secp256k1":
		privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
		if err != nil {
			return err
		}
		privBytes, err := marshalECPrivateKey(privateKey)
		if err != nil {
			return err
		}
		pubBytes, err := json.Marshal(privateKey.PublicKey)
		if err != nil {
			return err
		}
		address := crypto.PubkeyToAddress(privateKey.PublicKey)
		w.Address = address.Hex()
		secret["private_key"] = base64.RawStdEncoding.EncodeToString(privBytes)
		secret["public_key"] = base64.RawStdEncoding.EncodeToString(pubBytes)
	case "ed25519":
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}
		privateKeyString, err := getPemEncodedPrivateKey(privKey)
		if err != nil {
			return err
		}
		publicKeyString, err := getPemEncodedPublicKey(pubKey)
		if err != nil {
			return err
		}
		secret["private_key"] = privateKeyString
		secret["public_key"] = publicKeyString
	default:
		return fmt.Errorf("invalid algorithm")
	}
	if err := vault.AddSecret(ctx, w.Name, secret); err != nil {
		return err
	}
	return nil
}

func getPemEncodedPrivateKey(privKey interface{}) (string, error) {
	bytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return "", err
	}
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: bytes})
	return base64.RawStdEncoding.EncodeToString(pemEncoded), nil
}

func getPemEncodedPublicKey(pubKey interface{}) (string, error) {
	bytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", err
	}
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: bytes})
	return base64.RawStdEncoding.EncodeToString(pemEncodedPub), nil
}

func getDecodedPrivateKey(privKey string) (interface{}, error) {
	privBytes, err := base64.RawStdEncoding.DecodeString(privKey)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privBytes)
	x509Encoded := block.Bytes
	return x509.ParsePKCS8PrivateKey(x509Encoded)
}

func getDecodedPublicKey(publicKey string) (interface{}, error) {
	pubBytes, err := base64.RawStdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pubBytes)
	x509Encoded := block.Bytes
	return x509.ParsePKIXPublicKey(x509Encoded)
}

func marshalECPrivateKey(key *ecdsa.PrivateKey) ([]byte, error) {
	if !key.Curve.IsOnCurve(key.X, key.Y) {
		return nil, errors.New("invalid elliptic key public key")
	}
	privateKey := make([]byte, (key.Curve.Params().N.BitLen()+7)/8)
	return asn1.Marshal(ecPrivateKey{
		Version:    1,
		PrivateKey: key.D.FillBytes(privateKey),
		PublicKey:  asn1.BitString{Bytes: elliptic.Marshal(key.Curve, key.X, key.Y)},
	})
}

func parseECPrivateKey(der []byte) (key *ecdsa.PrivateKey, err error) {
	var privKey ecPrivateKey
	if _, err := asn1.Unmarshal(der, &privKey); err != nil {
		if _, err := asn1.Unmarshal(der, &pkcs8{}); err == nil {
			return nil, errors.New("x509: failed to parse private key (use ParsePKCS8PrivateKey instead for this key format)")
		}
		if _, err := asn1.Unmarshal(der, &pkcs1PrivateKey{}); err == nil {
			return nil, errors.New("x509: failed to parse private key (use ParsePKCS1PrivateKey instead for this key format)")
		}
		return nil, errors.New("x509: failed to parse EC private key: " + err.Error())
	}
	if privKey.Version != ecPrivKeyVersion {
		return nil, fmt.Errorf("x509: unknown EC private key version %d", privKey.Version)
	}

	var curve elliptic.Curve
	curve = secp256k1.S256()

	k := new(big.Int).SetBytes(privKey.PrivateKey)
	curveOrder := curve.Params().N
	if k.Cmp(curveOrder) >= 0 {
		return nil, errors.New("x509: invalid elliptic curve private key value")
	}
	priv := new(ecdsa.PrivateKey)
	priv.Curve = curve
	priv.D = k

	privateKey := make([]byte, (curveOrder.BitLen()+7)/8)

	// Some private keys have leading zero padding. This is invalid
	// according to [SEC1], but this code will ignore it.
	for len(privKey.PrivateKey) > len(privateKey) {
		if privKey.PrivateKey[0] != 0 {
			return nil, errors.New("x509: invalid private key length")
		}
		privKey.PrivateKey = privKey.PrivateKey[1:]
	}

	// Some private keys remove all leading zeros, this is also invalid
	// according to [SEC1] but since OpenSSL used to do this, we ignore
	// this too.
	copy(privateKey[len(privateKey)-len(privKey.PrivateKey):], privKey.PrivateKey)
	priv.X, priv.Y = curve.ScalarBaseMult(privateKey)

	return priv, nil
}
