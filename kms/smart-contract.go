package kms

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"wallet-kms/vault"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

type ContractMetadata struct {
	ABI     string
	Bin     string
	ChainId string
	Params  []interface{}
}

type SmartContract struct {
	SmartContractCaller     // Read-only binding to the contract
	SmartContractTransactor // Write-only binding to the contract
	SmartContractFilterer   // Log filterer for contract events
}

type SmartContractCaller struct {
	Contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

type SmartContractTransactor struct {
	Contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

type SmartContractFilterer struct {
	Contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

func DeploySmartContract(auth *bind.TransactOpts, backend bind.ContractBackend, cm ContractMetadata) (common.Address, *types.Transaction, *SmartContract, error) {

	var ContractData = &bind.MetaData{
		ABI: cm.ABI,
		Bin: cm.Bin,
	}

	parsed, err := ContractData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractData.Bin), backend, cm.Params...)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SmartContract{SmartContractCaller: SmartContractCaller{Contract: contract}, SmartContractTransactor: SmartContractTransactor{Contract: contract}, SmartContractFilterer: SmartContractFilterer{Contract: contract}}, nil
}

func bindSmartContract(address common.Address, ABI string, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

func NewSmartContract(address common.Address, ABI string, backend bind.ContractBackend) (*SmartContract, error) {
	contract, err := bindSmartContract(address, ABI, backend, backend, backend)
	if err != nil {
		return nil, err
	}

	return &SmartContract{SmartContractCaller: SmartContractCaller{Contract: contract}, SmartContractTransactor: SmartContractTransactor{Contract: contract}, SmartContractFilterer: SmartContractFilterer{Contract: contract}}, nil
}

func (w *Wallet) TransactionOptionsWithKMSSigning(ctx context.Context, vault vault.Vault, chainID *big.Int) (*bind.TransactOpts, error) {
	signer := types.LatestSignerForChainID(chainID)
	signerFn := func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		txHashBytes := signer.Hash(tx).Bytes()
		signature, err := w.SignTransactionHash(ctx, vault, txHashBytes)
		if err != nil {
			return nil, err
		}
		return tx.WithSignature(signer, signature)
	}
	return &bind.TransactOpts{
		Signer: signerFn,
	}, nil
}

func (w *Wallet) SignTransactionHash(ctx context.Context, vault vault.Vault, transactionHash []byte) ([]byte, error) {
	var signature []byte
	data, err := vault.GetSecret(ctx, w.Name)
	if err != nil {
		return nil, err
	}
	privateKeyString := data["private_key"].(string)
	switch w.Algorithm {
	case "secp256k1":
		var privKey *ecdsa.PrivateKey
		privKeyBytes, err := base64.RawStdEncoding.DecodeString(privateKeyString)
		if err != nil {
			return nil, err
		}
		privKey, err = parseECPrivateKey(privKeyBytes)
		if err != nil {
			return nil, err
		}
		signature, err = secp256k1.Sign(transactionHash, math.PaddedBigBytes(privKey.D, 32))
		if err != nil {
			return nil, err
		}
	case "ed25519":
		privKey, err := getDecodedPrivateKey(privateKeyString)
		if err != nil {
			return nil, err
		}
		signature = ed25519.Sign(privKey.(ed25519.PrivateKey), transactionHash)
	default:
		return nil, fmt.Errorf("invalid algorithm")
	}
	return signature, nil
}
