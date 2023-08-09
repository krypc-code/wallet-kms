# Self Managed Wallets

The Self-Managed Wallet (SMW) system is designed as a custodial wallet solution that operates within the client's environment. It empowers users to generate and retain cryptographic keys within their infrastructure, ensuring a higher level of control and security. The architecture incorporates a robust key storage mechanism to safeguard keys, with the added convenience of integrating with KrypCore Web3 services, enabling seamless utilization of KrypCore functionalities using self-generated keys. 
## Running Self Managed Wallets

Follow these steps to set up and run a self-managed wallet using the provided repository.

### Cloning the Repository

Start by cloning the repository to your local system using the following command:

```bash
git clone git@github.com:krypc-code/wallet-kms.git
```

### Setting Up the Environment

After cloning the repository, navigate to the `test` folder:

```bash
cd test
```

Launch the HashiCorp Vault service using Docker Compose:

```bash
sudo docker-compose -f docker-compose-vault.yaml up -d
```

Once the Vault service is running, create secrets keys within the Vault and update the environment variables in the `docker-compose-kms.yaml` file as follows:

```yaml
"VAULT_URL": "http://127.0.0.1:8200",
"VAULT_TOKEN": "hvs.6xYh1CwO2ekf1ZpjryumvqmQ",
"AUTH_TOKEN": "abd3789a-8a15-4e5a-8644-ed65a2c2e7f6",
"PROXY_URL": "http://localhost:8888",
"ENDPOINT": "https://polygon-mumbai-dev-node.krypcore.com/api/v0/rpc?apiKey=1ddc4575-fe65-4f00-a420-9d8a7a4086aa&token=abd3789a-8a15-4e5a-8644-ed65a2c2e7f6",
"WALLET_INSTANCE_ID": "INS_NC_17_2023721",
"SUBSCRIPTION_ID": "6186952413",
"SCHEDULER_DURATION": "10"
```

### Running the Service

Once you've configured the environment variables, run the self-managed wallet service using the following command:

```bash
sudo docker-compose -f docker-compose-kms.yaml up -d
```

### Creating a Wallet

After the self-managed wallet service is up and running, use the following curl command to create a wallet:

```bash
curl -d '{"name":"wallet2", "algorithm": "secp256k1"}' -H "Content-Type: application/json" -X POST http://localhost:8889/wallet/createWallet
```

You will receive a response with a unique wallet ID, which you can use for further operations.

### Submitting a Transaction

To submit a transaction, use the following curl command:

```bash
curl -d '{
  "walletId": "effae2b6-3ee3-48cb-9528-87c29152c89e",
  "to": "0xc2de797fab7d2d2b26246e93fcf2cd5873a90b10",
  "chainId": 80001,
  "method": "store",
  "params": [{"type": "uint256", "value": 35}],
  "isContractTxn": true,
  "contractABI": "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"num\",\"type\":\"uint256\"}],\"name\":\"store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"retrieve\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"
}' -H "Content-Type: application/json" -H "instanceId: INS_WA_4_2023630" -H "Authorization: 49817601-81d6-4863-bc70-619211d35efd_dcd96173-bd93-46ff-8d3a-a98baeb4fa87" -X POST http://localhost:8889/wallet/submitTransaction
```

### Deploying a Smart Contract

To deploy a smart contract, use the following curl command:

```bash
curl -d '{
  "walletId": "effae2b6-3ee3-48cb-9528-87c29152c89e",
  "byteCode": "",
  "abi": "",
  "params": []
}' -H "Content-Type: application/json" -H "instanceId: INS_WA_4_2023630" -H "Authorization: 49817601-81d6-4863-bc70-619211d35efd_dcd96173-bd93-46ff-8d3a-a98baeb4fa87" -X POST http://localhost:8889/wallet/deployContract
```

This will allow you to deploy a smart contract and receive a transaction hash in response.

Remember to replace the placeholders with actual values as needed.