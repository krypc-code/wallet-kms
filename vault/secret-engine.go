package vault

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

const (
	DefaultMountPath = "secret"
	ServiceName      = "NC-WALLET"
)

func (vault *HashiCorp) AddSecret(ctx context.Context, secretKey string, data map[string]interface{}) error {
	log.Println("Vault---AddVaultSecret()---STARTS", time.Now())
	defer log.Println("Vault---AddVaultSecret()---ENDS", time.Now())
	if data != nil {
		secretPath := ServiceName + "/" + secretKey
		secret, err := vault.client.KVv2(DefaultMountPath).Put(ctx, secretPath, data)
		if err != nil {
			return fmt.Errorf("unable to write secret: %s", err)
		}
		log.Println("secret added successfully.", secret.Raw.RequestID)
		return nil
	} else {
		return fmt.Errorf("empty data")
	}
}

func (vault *HashiCorp) GetSecret(ctx context.Context, secretKey string) (map[string]interface{}, error) {
	log.Println("Vault---retrieveVaultSecret()---STARTS", time.Now())
	defer log.Println("Vault---retrieveVaultSecret()---ENDS", time.Now())
	if secretKey != "" {
		secretPath := ServiceName + "/" + secretKey
		secret, err := vault.client.KVv2(DefaultMountPath).Get(ctx, secretPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read secret: %s", err.Error())
		}
		if secret.Data != nil {
			return secret.Data, err
		} else {
			return nil, errors.New("no secret found")
		}
	} else {
		return nil, fmt.Errorf("empty secret path")
	}
}

func (vault *HashiCorp) DeleteSecret(ctx context.Context, secretKey string) error {
	log.Println("Vault---DeleteVaultSecret()---STARTS", time.Now())
	defer log.Println("Vault---DeleteVaultSecret()---ENDS", time.Now())
	if secretKey != "" {
		secretPath := ServiceName + "/" + secretKey
		log.Println("delete request for secret path", secretPath)
		err := vault.client.KVv2(DefaultMountPath).Delete(ctx, secretPath)
		if err != nil {
			return fmt.Errorf("unable to delete secret %s", err)
		}
		log.Println("requested secret deleted successfully.")
		return nil
	} else {
		return fmt.Errorf("empty secret path")
	}
}
