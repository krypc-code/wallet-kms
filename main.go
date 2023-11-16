package main

import "wallet-kms/kms"

// @title KMS Wallet
// @version 1.0
// @description KMS Wallet Implementation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8889
// @BasePath /wallet
// @schemes https
func main() {
	kms.Run()
}
