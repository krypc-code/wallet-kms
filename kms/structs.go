package kms

const (
	NAMESPACE = "wallet"
)

type ResponseBody struct {
	Status  string
	Message string
	Data    interface{}
}

type WalletRequest struct {
	Algorithm string
}

type WalletResponse struct {
	WalletId string
}
