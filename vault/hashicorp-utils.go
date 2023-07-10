package vault

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"time"

	vault "github.com/hashicorp/vault/api"
)

func getVaultClient(sourceUrl, token string) (*vault.Client, error) {
	log.Println("vault----loadVaultConfig()--STARTS", time.Now())
	defer log.Println("vault----loadVaultConfig()--ENDS", time.Now())
	if sourceUrl != "" && token != "" {
		log.Printf("input params source url %v and token is not empty", sourceUrl)
		isConnected := CheckInstanceUp(sourceUrl)
		if isConnected {
			isUnseal, err := isVaultUnsealed(sourceUrl)
			if err != nil {
				return nil, fmt.Errorf("error in vault un seal check %s", err)
			}
			if isUnseal {
				log.Println("New instance", time.Now())
				config := vault.DefaultConfig()
				config.Address = sourceUrl
				vaultClient, err := vault.NewClient(config)
				if err != nil {
					return nil, fmt.Errorf("unable to initialize Vault client %s", err)
				}
				vaultClient.SetToken(token)
				return vaultClient, nil
			} else {
				return nil, fmt.Errorf("vault is not ready - unseal is pending %s", err)
			}
		} else {
			log.Println("dial connection timed out")
			return nil, fmt.Errorf("dial connection timed out")
		}
	} else {
		return nil, fmt.Errorf("mandatory values are missing")
	}
}

const unsealStatusCheck = "/v1/sys/seal-status"

func CheckInstanceUp(url string) bool {
	log.Println("utils -- CheckInstanceUp() -- STARTS", time.Now())
	client := http.Client{
		Timeout: time.Second * 5,
	}
	log.Println("utils -- CheckInstanceUp() -- URL", url)
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error connecting to the instance:", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return true
	}
	defer log.Println("utils -- CheckInstanceUp() -- STARTS", time.Now())
	return false
}

func isVaultUnsealed(url string) (bool, error) {
	resp, err := http.Get(url + unsealStatusCheck)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	// Check if the response indicates that Vault is unsealed
	if resp.StatusCode == http.StatusOK {
		// Parse the response JSON and check if the 'sealed' field is false
		unsealed, err := isUnsealed(body)
		if err != nil {
			return false, err
		}
		return unsealed, nil
	}
	return false, nil
}
func isEmpty(config interface{}) bool {
	v := reflect.ValueOf(config)
	return v.Kind() == reflect.Struct && !v.IsZero()
}

func isUnsealed(responseBody []byte) (bool, error) {
	// Parse the response JSON
	// Assuming the JSON has a 'sealed' field indicating whether Vault is sealed or unsealed
	// Modify the parsing logic according to the actual response structure
	type SealStatus struct {
		Sealed bool `json:"sealed"`
	}
	var status SealStatus
	err := json.Unmarshal(responseBody, &status)
	if err != nil {
		return false, err
	}
	return !status.Sealed, nil
}
