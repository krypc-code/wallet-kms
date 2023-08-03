package utils

const (
	NAMESPACE = "wallet"
)

type Config struct {
	AuthToken      string
	ProxyUrl       string
	Endpoint       string
	InstanceId     string
	SubscriptionId string
}

type WalletRequest struct {
	Name      string `json:"name"`
	Algorithm string `json:"algorithm"`
}

type WalletResponse struct {
	WalletId string
}

type TransactionResponse struct {
	ApiKey      string
	WalletId    string
	ProxyUrl    string
	Transaction string
}

type DeployContractRequest struct {
	WalletId string  `json:"walletId"`
	ByteCode string  `json:"byteCode"`
	ABI      string  `json:"abi"`
	Params   []Param `json:"params"`
}

type DeployContractResponse struct {
	TxnHash         string `json:"txHash,omitempty"`
	ContractAddress string `json:"contractAddress,omitempty"`
}

type SignAndSubmitTxn struct {
	WalletId string  `json:"walletId"`
	To       string  `json:"to"`
	Gas      uint64  `json:"gas"`
	Value    int64   `json:"value"`
	Method   string  `json:"method"`
	Params   []Param `json:"params"`
	// Params        []interface{} `json:"params"`
	IsContractTxn bool   `json:"isContractTxn"`
	ContractABI   string `json:"contractABI"`
	Data          string `json:"data"`
}

type EstimateGasRequest struct {
	WalletId      string  `json:"walletId"`
	To            string  `json:"to,omitempty" example:"0xc2de797fab7d2d2b26246e93fcf2cd5873a90b10"`
	Gas           uint64  `json:"gas,omitempty"`
	Value         int64   `json:"value,omitempty"`
	Method        string  `json:"method,omitempty" example:"store"`
	Params        []Param `json:"params,omitempty" `
	IsContractTxn bool    `json:"isContractTxn,omitempty" example:"true"`
	ContractABI   string  `json:"contractABI,omitempty" example:""`
	Data          string  `json:"data,omitempty" example:""`
	ByteCode      string  `json:"byteCode,omitempty" example:""`
	Service       string  `json:"service,omitempty" example:""`
}

type EstimatedGasResponse struct {
	Address      string `json:"address"`
	EstimatedGas uint64 `json:"estimatedGas"`
}

type Param struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type SignAndSubmitTxnResponse struct {
	TxnHash string `json:"txHash"`
}

type AddWalletRequest struct {
	WalletId  string `json:"walletId"`
	Address   string `json:"address,omitempty"`
	Name      string `json:"name"`
	Algorithm string `json:"algorithm,omitempty"`
}

type NonceRequest struct {
	WalletId        string `json:"walletId"`
	ChainId         string `json:"chainId"`
	ReferenceId     string `json:"referenceId"`
	Type            string `json:"type"`
	TxnHash         string `json:"txnHash"`
	ContractAddress string `json:"contractAddress"`
}

// type GetNcNonceResponse struct {
// 	WalletId string `json:"walletId,omitempty"`
// 	Nonce    uint64 `json:"nonce,omitempty"`
// }

type Deploy struct {
	ReferenceId    string  `json:"referenceId"`
	SubscriptionId string  `json:"subscriptionId"`
	WalletId       string  `json:"walletId"`
	InstanceId     string  `json:"instanceId"`
	ByteCode       string  `json:"byteCode"`
	ABI            string  `json:"abi"`
	Params         []Param `json:"params"`
}

type Transaction struct {
	ReferenceId    string  `json:"referenceId" bson:"reference_id"`
	SubscriptionId string  `json:"subscriptionId" bson:"subscription_id"`
	WalletId       string  `json:"walletId" bson:"wallet_id"`
	InstanceId     string  `json:"instanceId" bson:"instance_id"`
	To             string  `json:"to" bson:"to"`
	Gas            uint64  `json:"gas" bson:"gas"`
	Value          int64   `json:"value" bson:"value"`
	Method         string  `json:"method" bson:"method"`
	Params         []Param `json:"params" bson:"params"`
	IsContractTxn  bool    `json:"isContractTxn" bson:"is_contract_txn"`
	ContractABI    string  `json:"contractABI" bson:"contract_abi"`
	Data           string  `json:"data" bson:"data"`
}
