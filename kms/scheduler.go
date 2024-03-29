package kms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"os"
	"strconv"
	"time"
	"wallet-kms/utils"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

func (s *Service) ScheduleService() error {
	ctx := context.Background()
	duration, err := strconv.Atoi(os.Getenv("SCHEDULER_DURATION"))
	if err != nil {
		return err
	}
	for {
		time.Sleep(time.Duration(duration) * time.Second)
		s.e.Logger.Infof("fetching records")
		s.executor.Publish(s.DeployContracts, ctx)
		s.executor.Publish(s.SubmitTransactions, ctx)
		s.executor.Publish(s.ApprovePendingWallets, ctx)
	}
}

func (s *Service) DeployContracts(ctx context.Context) {
	records, err := utils.GetAllDeployRecords(s.config)
	if err != nil {
		s.e.Logger.Errorf(err.Error())
	}
	for _, record := range records {
		walletId, err := uuid.Parse(record.WalletId)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		walletBytes, err := s.db.Get([]byte(utils.NAMESPACE), walletId.NodeID())
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		wallet := &Wallet{}
		if err := json.Unmarshal(walletBytes, wallet); err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		client, err := utils.GetEthereumClient(ctx, s.config)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		chainId, err := client.ChainID(ctx)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		rawDecodedText, err := base64.StdEncoding.DecodeString(record.ABI)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		nonce, err := utils.GetNonceFromPlatform(s.config, &utils.NonceRequest{WalletId: wallet.WalletId, ChainId: chainId.String()})
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		transactOpts, err := TransactionOptionsWithKMSSigning(ctx, wallet, s.vault, chainId)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		transactOpts.Nonce = big.NewInt(int64(nonce))
		tipCap, _ := client.SuggestGasTipCap(ctx)
		feeCap, _ := client.SuggestGasPrice(ctx)
		transactOpts.GasFeeCap = feeCap
		transactOpts.GasTipCap = tipCap
		transactOpts.GasLimit = 5000000 // Todo Need an attention
		contractMeta := ContractMetadata{
			ABI:     string(rawDecodedText),
			Bin:     record.ByteCode,
			ChainId: chainId.String(),
		}
		params, err := utils.ConvertParamsAsPerTypes(record.Params)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		contractMeta.Params = params
		address, txn, _, err := DeploySmartContract(transactOpts, client, contractMeta)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		err = utils.UpdatePlatformNonce(s.config, &utils.NonceRequest{WalletId: wallet.WalletId, ChainId: chainId.String(), TxnHash: txn.Hash().String(),
			ContractAddress: address.String(), Type: "deploy", ReferenceId: record.ReferenceId})
		if err != nil {
			s.e.Logger.Errorf(err.Error())
		}
	}
}

func (s *Service) SubmitTransactions(ctx context.Context) {
	records, err := utils.GetAllTransactionRecords(s.config)
	if err != nil {
		s.e.Logger.Errorf(err.Error())
	}
	for _, record := range records {
		walletId, err := uuid.Parse(record.WalletId)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		walletBytes, err := s.db.Get([]byte(utils.NAMESPACE), walletId.NodeID())
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		wallet := &Wallet{}
		if err := json.Unmarshal(walletBytes, wallet); err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		var to common.Address
		if record.To != "" {
			to = common.HexToAddress(record.To)
		}
		client, err := utils.GetEthereumClient(ctx, s.config)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		chainId, err := client.ChainID(ctx)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		nonce, err := utils.GetNonceFromPlatform(s.config, &utils.NonceRequest{WalletId: wallet.WalletId, ChainId: chainId.String()})
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		var txnHash string
		if record.IsContractTxn {
			var contractABI string
			if record.ContractABI != "" {
				contractABI = record.ContractABI
			} else {
				s.e.Logger.Errorf(err.Error())
				continue
			}
			contract, err := NewSmartContract(to, contractABI, client)
			if err != nil {
				s.e.Logger.Errorf(err.Error())
				continue
			}
			txnOpts, err := TransactionOptionsWithKMSSigning(ctx, wallet, s.vault, chainId)
			if err != nil {
				s.e.Logger.Errorf(err.Error())
				continue
			}
			if record.Value > 0 {
				txnOpts.Value = big.NewInt(record.Value)
			}
			txnOpts.Nonce = big.NewInt(int64(nonce))
			params, err := utils.ConvertParamsAsPerTypes(record.Params)
			if err != nil {
				s.e.Logger.Errorf(err.Error())
				continue
			}
			txn, err := contract.SmartContractTransactor.Contract.Transact(txnOpts, record.Method, params...)
			if err != nil {
				s.e.Logger.Errorf(err.Error())
				continue
			}
			txnHash = txn.Hash().String()
		} else {
			gasPrice, err := client.SuggestGasPrice(ctx)
			if err != nil {
				s.e.Logger.Errorf(err.Error())
			}
			var dataBytes []byte
			if record.Data != "" {
				dataBytes, err = base64.RawStdEncoding.DecodeString(record.Data)
				if err != nil {
					s.e.Logger.Errorf(err.Error())
					continue
				}
			}
			var estimatedGas uint64
			if record.Gas > 0 {
				estimatedGas = record.Gas
			} else {
				estimatedGas, err = client.EstimateGas(ctx, ethereum.CallMsg{From: common.HexToAddress(wallet.Address), To: &to, Value: big.NewInt(record.Value), Data: dataBytes, GasPrice: gasPrice})
				if err != nil {
					s.e.Logger.Errorf(err.Error())
					continue
				}
			}
			txn := types.NewTx(&types.LegacyTx{
				Nonce:    nonce,
				GasPrice: gasPrice,
				Gas:      estimatedGas,
				Value:    big.NewInt(record.Value),
				Data:     dataBytes,
				To:       &to,
			})
			signer := types.LatestSignerForChainID(chainId)
			signature, err := SignTransactionHash(ctx, wallet, s.vault, txn.Hash().Bytes())
			txn, err = txn.WithSignature(signer, signature)
			if err != nil {
				s.e.Logger.Errorf(err.Error())
				continue
			}
			if err := client.SendTransaction(ctx, txn); err != nil {
				s.e.Logger.Errorf(err.Error())
				continue
			}
			txnHash = txn.Hash().String()
		}
		err = utils.UpdatePlatformNonce(s.config, &utils.NonceRequest{WalletId: wallet.WalletId, ChainId: chainId.String(), TxnHash: txnHash,
			Type: "txn", ReferenceId: record.ReferenceId})
		if err != nil {
			s.e.Logger.Errorf(err.Error())
		}

	}
}

func (s *Service) ApprovePendingWallets(ctx context.Context) {
	records, err := utils.GetPendingWalletRecords(s.config)
	if err != nil {
		s.e.Logger.Errorf(err.Error())
	}
	for _, record := range records {
		walletId := uuid.New()
		wallet := Wallet{
			Name:      record.WalletName,
			Algorithm: record.Algorithm,
			WalletId:  walletId.String(),
		}
		if err := wallet.generateKey(ctx, s.vault); err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		data, err := json.Marshal(wallet)
		if err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		if err := s.db.Set([]byte(utils.NAMESPACE), walletId.NodeID(), data); err != nil {
			s.e.Logger.Errorf(err.Error())
			continue
		}
		if err := utils.UpdatePendingWallet(s.config, &utils.AddWalletRequest{
			WalletId:  wallet.WalletId,
			Address:   wallet.Address,
			Name:      wallet.Name,
			Algorithm: wallet.Algorithm,
		}); err != nil {
			s.e.Logger.Errorf(err.Error())
		}
	}
}
