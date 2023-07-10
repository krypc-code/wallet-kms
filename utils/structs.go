package utils

const (
	NAMESPACE = "wallet"
)

type Config struct {
	AuthToken  string
	ProxyUrl   string
	Endpoint   string
	InstanceId string
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
	WalletId string        `json:"walletId"`
	ByteCode string        `json:"byteCode"`
	ABI      string        `json:"abi"`
	Params   []interface{} `json:"params"`
	ChainId  int64         `json:"chainId"`
}

type DeployContractResponse struct {
	TxnHash         string `json:"txHash,omitempty"`
	ContractAddress string `json:"contractAddress,omitempty"`
}

type SignAndSubmitTxn struct {
	WalletId string  `json:"walletId"`
	ChainID  uint64  `json:"chainId,omitempty"`
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

type Param struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
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
	WalletId string `json:"walletId,omitempty"`
	ChainId  string `json:"chainId"`
}

type GetNcNonceResponse struct {
	WalletId string `json:"walletId,omitempty"`
	Nonce    uint64 `json:"nonce,omitempty"`
}
