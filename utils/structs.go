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
	Address  string
}

type WalletBalanceRequest struct {
	WalletId string `json:"walletId"`
	ChainId  string `json:"chainId"`
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
	GasLimit uint64  `json:"gasLimit"`
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
	To            string  `json:"to,omitempty"`
	Gas           uint64  `json:"gas,omitempty"`
	Value         int64   `json:"value,omitempty"`
	Method        string  `json:"method,omitempty"`
	Params        []Param `json:"params,omitempty" `
	IsContractTxn bool    `json:"isContractTxn,omitempty"`
	ContractABI   string  `json:"contractABI,omitempty"`
	Data          string  `json:"data,omitempty"`
	ByteCode      string  `json:"byteCode,omitempty"`
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
	Gas            uint64  `json:"gas,omitempty"`
}

type Transaction struct {
	ReferenceId    string  `json:"referenceId"`
	SubscriptionId string  `json:"subscriptionId"`
	WalletId       string  `json:"walletId"`
	InstanceId     string  `json:"instanceId"`
	To             string  `json:"to"`
	Gas            uint64  `json:"gas"`
	Value          int64   `json:"value"`
	Method         string  `json:"method"`
	Params         []Param `json:"params"`
	IsContractTxn  bool    `json:"isContractTxn"`
	ContractABI    string  `json:"contractABI"`
	Data           string  `json:"data"`
}

type CallContractRequest struct {
	WalletId    string  `json:"walletId"`
	To          string  `json:"to,omitempty" example:"0xc2de797fab7d2d2b26246e93fcf2cd5873a90b10"`
	Gas         uint64  `json:"gas,omitempty"`
	Value       int64   `json:"value,omitempty"`
	Method      string  `json:"method,omitempty" example:"store"`
	Params      []Param `json:"params,omitempty" `
	ContractABI string  `json:"contractABI,omitempty" example:"hello"`
}

type CallContractResponse struct {
	Response *[]interface{} `json:"response"`
}

type EIP712SignRequest struct {
	WalletId string `json:"walletId"`
	Data     string `json:"data"`
}

type SignMsgRequest struct {
	WalletId string `json:"walletId"`
	Message  string `json:"message"`
}

type VerifyMsgRequest struct {
	WalletId  string `json:"walletId"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type VerifyMsgResponse struct {
	IsVerified bool `json:"isVerified" example:"true"`
}

type BalanceRequest struct {
	Address string `json:"address"`
	ChainId string `json:"chainId"`
}

type BalanceResponse struct {
	Address string `json:"address"`
	Balance uint64 `json:"balance"`
}

type PendingWalletResponse struct {
	UniqueId   string `json:"uniqueId"`
	InstanceId string `json:"instanceId"`
	WalletName string `json:"walletName"`
	Algorithm  string `json:"algorithm"`
}

type SignAndSubmitGSNTxnRequest struct {
	WalletId    string        `json:"walletId,omitempty"`
	ChainID     uint64        `json:"chainId,omitempty" example:"80001"`
	DAppId      string        `json:"dAppId,omitempty"`
	To          string        `json:"to,omitempty" example:"store"`
	Gas         uint64        `json:"gas,omitempty" `
	Value       int64         `json:"value,omitempty" `
	Method      string        `json:"method,omitempty" example:"store"`
	Params      []interface{} `json:"params,omitempty"`
	ContractABI string        `json:"contractABI,omitempty" example:""`
}

type GSNTxnPayloadRequest struct {
	ChainId         uint64        `json:"chainId,omitempty"`
	DAppId          string        `json:"dAppId,omitempty"`
	UserAddress     string        `json:"userAddress,omitempty"`
	ContractAddress string        `json:"contractAddress,omitempty"`
	ContractAbi     string        `json:"contractAbi,omitempty"`
	Method          string        `json:"method,omitempty"`
	Args            []interface{} `json:"args,omitempty"`
}

type GSNSendTxnRequest struct {
	ChainId         uint64      `json:"chainId,omitempty"`
	DAppId          string      `json:"dAppId,omitempty"`
	UserAddress     string      `json:"userAddress,omitempty"`
	ContractAddress string      `json:"contractAddress,omitempty"`
	Method          string      `json:"method,omitempty"`
	Request         interface{} `json:"request,omitempty"`
	Signature       string      `json:"signature,omitempty"`
	DomainSeparator string      `json:"domainSeparator,omitempty"`
	SignatureType   string      `json:"signatureType,omitempty"`
}
