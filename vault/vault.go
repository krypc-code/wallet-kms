package vault

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

type Vault interface {
	GetPublicKey(keyName string) (*ecdsa.PublicKey, error)
	GenerateKey(keyName, algorithm string) (*vault.Secret, error)
	SignTransactionHash(keyName string, transactionHash []byte) ([]byte, error)
	AddSecret(ctx context.Context, secretKey string, data map[string]interface{}) error
	GetSecret(ctx context.Context, secretKey string) (map[string]interface{}, error)
	DeleteSecret(ctx context.Context, secretPath string) error
}

type HashiCorp struct {
	client *vault.Client
}

func NewHashiCorpVault(url, token string) (Vault, error) {
	client, err := getVaultClient(url, token)
	if err != nil {
		return nil, err
	}
	vault := &HashiCorp{
		client: client,
	}
	return vault, nil
}

func (vault *HashiCorp) GenerateKey(keyName, algorithm string) (*vault.Secret, error) {
	response, err := vault.client.Logical().Write(fmt.Sprintf("transit/keys/%s", keyName), map[string]interface{}{
		"type": algorithm,
	})
	return response, err
}

func (vault *HashiCorp) GetPublicKey(keyName string) (*ecdsa.PublicKey, error) {
	response, err := vault.client.Logical().Read(fmt.Sprintf("transit/keys/%s", keyName))
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("key not found")
	}
	publicKeyPem := response.Data["keys"].(map[string]interface{})["1"].(map[string]interface{})["public_key"].(string)
	block, _ := pem.Decode([]byte(publicKeyPem))
	if block == nil {
		return nil, fmt.Errorf("Failed to decode public key")
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Failed to convert public key to ECDSA")
	}
	return ecdsaPublicKey, nil
}

func (vault *HashiCorp) SignTransactionHash(keyName string, transactionHash []byte) ([]byte, error) {
	signatureData := map[string]interface{}{
		"hash_input": hex.EncodeToString(transactionHash),
	}
	response, err := vault.client.Logical().Write(fmt.Sprintf("transit/sign/%s", keyName), signatureData)
	if err != nil {
		return nil, err
	}
	signatureHex := response.Data["signature"].(string)
	sig := strings.Split(signatureHex, ":")
	return []byte(sig[2]), nil
}
