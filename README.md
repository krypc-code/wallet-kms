# Self Managed Wallets

The Self-Managed Wallet (SMW) system is designed as a custodial wallet solution that operates within the client's environment. It empowers users to generate and retain cryptographic keys within their infrastructure, ensuring a higher level of control and security. The architecture incorporates a robust key storage mechanism to safeguard keys, with the added convenience of integrating with KrypCore Web3 services, enabling seamless utilization of KrypCore functionalities using self-generated keys. 


## Prerequisite

Docker
Docker Compose

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

**Vault Initialization Steps:**

1. **Access the Vault UI:**
    - Start the Vault server.
    - Open a web browser and navigate to http://127.0.0.1:8200.
   
      ![](docs/assets/initialize.png)

2. **Set Keyshare and Key Threshold:**
    - In the Vault UI, locate the settings for key sharing.
    - Set the number of keyshares (e.g., 5) and the key threshold (e.g., 3).
    - These values determine the number of key parts required to unseal the vault.
   

3. **Generate Key Pairs:**
    - Click on the "Initialize" button in the Vault UI.
    - This action triggers the generation of a set of key and root token pairs.
      ![](docs/assets/keys.png)

4. **Download and Save Keys:**
    - After initialization, download the generated keys.
    - Save the downloaded keys securely on your local machine.
      ![](docs/assets/downloadkey.png)

5. **Provide Unseal Keys:**
    - Open the downloaded key file and find the keys_base64 values.
    - Depending on the threshold set earlier, gather the required number of keys_base64 values.
    - These keys will be used to unseal the vault.

6. **Unseal the Vault:**
    - In the Vault UI, locate the "Unseal" section.
    - Paste the collected keys_base64 values into the designated fields, based on the threshold.
    - Click "Continue" to unseal the vault.
      ![](docs/assets/unseal.png)

7. **Provide Root Token:**
    - In the Vault UI, find the field to input the root token.
    - Enter the root token obtained during the initialization process.
      ![](docs/assets/signin.png)

8. **Sign In to Vault:**
    - Click on the "Sign In" or "Log In" button in the Vault UI.
    - If the root token is valid, you will gain access to the Vault.
     ![](docs/assets/Login.png)

9. **Create New Engine:**
    - Create new secret engine by selecting KV on the options screen.
      ![](docs/assets/kv.png)

10. **Secret As Path:**
   - Create new path with value "secret" in path param.
      ![](docs/assets/secret.png)

11. **Vault Initialization Complete:**
    - At this point, your Vault is initialized and accessible.


Ensure that you keep the downloaded keys and root token secure.



Once the Vault service is running and initialized successfully,update the environment variables in the `docker-compose-kms.yaml` file as follows:

```yaml
"VAULT_URL": "http://127.0.0.1:8200",
"VAULT_TOKEN": "hvs.xxxxxxxxxxxxxxxxx",
"AUTH_TOKEN": "abd3789a-xxxx-xxxx-xxxx-ed65a2c2e7f6",
"PROXY_URL": "http://localhost:8888",
"ENDPOINT": "https://polygon-mumbai-dev-node.krypcore.com/api/v0/rpc?apiKey=1ddc4575-xxxx-xxxx-xxxx-9d8a7a4086aa&token=abd3789a-xxxx-xxxx-xxxx-ed65a2c2e7f6",
"WALLET_INSTANCE_ID": "XXX_XX_XX_2023721",
"SUBSCRIPTION_ID": "XXXXXXXXXX",
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
  "params": [{"type": "uint256", "value": "35"}],
  "isContractTxn": true,
  "contractABI": "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"num\",\"type\":\"uint256\"}],\"name\":\"store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"retrieve\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"
}' -H "Content-Type: application/json" -X POST http://localhost:8889/wallet/submitTransaction
```

### Deploying a Smart Contract

To deploy a smart contract, use the following curl command:

```bash
curl -d '{
  "walletId": "effae2b6-3ee3-48cb-9528-87c29152c89e",
  "byteCode": "",
  "abi": "",
  "params": []
}' -H "Content-Type: application/json" -X POST http://localhost:8889/wallet/deployContract
```

This will allow you to deploy a smart contract and receive a transaction hash and contract address in response.


### Estimating Gas Price

To estimate gas for a transaction, use the following curl command:

```bash
curl -d '{
  "walletId": "effae2b6-3ee3-48cb-9528-87c29152c89e",
  "to": "0xc2de797fab7d2d2b26246e93fcf2cd5873a90b10",
  "method": "store",
  "params": [{"type": "uint256", "value": "35"}],
  "isContractTxn": true,
  "contractABI": "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"num\",\"type\":\"uint256\"}],\"name\":\"store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"retrieve\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"
}' -H "Content-Type: application/json" -X POST http://localhost:8889/wallet/estimateGas
```

### Get Wallet Balance

To get balance of a wallet, use the following curl command:

```bash
curl -d '{
  "walletId": "effae2b6-3ee3-48cb-9528-87c29152c89e","chainId": "80001"
}' -H "Content-Type: application/json" -X POST http://localhost:8889/wallet/getBalance
```

### Call Contract Method

To call contract method, use the following curl command:

```bash
curl -d '{
  "walletId": "effae2b6-3ee3-48cb-9528-87c29152c89e",
  "to": "0xc2de797fab7d2d2b26246e93fcf2cd5873a90b10",
  "method": "store",
  "params": [{"type": "uint256", "value": "35"}],
  "contractABI": "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"num\",\"type\":\"uint256\"}],\"name\":\"store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"retrieve\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"
}' -H "Content-Type: application/json" -X POST http://localhost:8889/wallet/callContract
```

### Sign Message

To sign a message, use the following curl command:

```bash
curl -d '{"walletId": "effae2b6-3ee3-48cb-9528-87c29152c89e","message":"Hello"}' -H "Content-Type: application/json" -X POST http://localhost:8889/wallet/signMessage
```

### Verify Signature Offchain

To verify signature offchain, use the following curl command:

```bash
curl -d '{"walletId": "effae2b6-3ee3-48cb-9528-87c29152c89e","message":"Hello","signature":"0x0274ba1a35dd8dfcf279a660f970985036c1432ceead1e05b81443b9d94bac403e4e2e8dbab494fe428e212ed0e9b2f8ebac327c5971dc461c9b147bc33fbc5301"}' -H "Content-Type: application/json" -X POST http://localhost:8889/wallet/verifySignatureOffChain
```

### Sign and Submit Gasless Transaction

To Sign and Submit Gasless Transaction, use the following curl command:

```bash
curl --location 'http://localhost:8889/wallet/signAndSubmitGaslessTxn' \
--header 'Authorization: fa4b9b58-57e9-4772-af79-6353f61dd78b' \
--header 'Content-Type: application/json' \
--data '{
    "walletId":"aaad6fb4-0d6d-4d9d-8558-e19cbd4bdc7c",
    "dAppId":"DEV_TEST_XENO_188_20230912",
    "chainId": 80001,
    "to": "0x0E762313219aE4dD7C674367a39901Ac1c28Cef3",
    "userId":"DEV_XENO_80_20230912",
    "contractAbi": "W3siaW5wdXRzIjpbeyJpbnRlcm5hbFR5cGUiOiJzdHJpbmciLCJuYW1lIjoiY29sbGVjdGlvbk5hbWUiLCJ0eXBlIjoic3RyaW5nIn0seyJpbnRlcm5hbFR5cGUiOiJzdHJpbmciLCJuYW1lIjoic3ltYm9sIiwidHlwZSI6InN0cmluZyJ9LHsiaW50ZXJuYWxUeXBlIjoiYWRkcmVzcyIsIm5hbWUiOiJmb3J3YXJkZXIiLCJ0eXBlIjoiYWRkcmVzcyJ9XSwic3RhdGVNdXRhYmlsaXR5Ijoibm9ucGF5YWJsZSIsInR5cGUiOiJjb25zdHJ1Y3RvciJ9LHsiYW5vbnltb3VzIjpmYWxzZSwiaW5wdXRzIjpbeyJpbmRleGVkIjp0cnVlLCJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6Im93bmVyIiwidHlwZSI6ImFkZHJlc3MifSx7ImluZGV4ZWQiOnRydWUsImludGVybmFsVHlwZSI6ImFkZHJlc3MiLCJuYW1lIjoiYXBwcm92ZWQiLCJ0eXBlIjoiYWRkcmVzcyJ9LHsiaW5kZXhlZCI6dHJ1ZSwiaW50ZXJuYWxUeXBlIjoidWludDI1NiIsIm5hbWUiOiJ0b2tlbklkIiwidHlwZSI6InVpbnQyNTYifV0sIm5hbWUiOiJBcHByb3ZhbCIsInR5cGUiOiJldmVudCJ9LHsiYW5vbnltb3VzIjpmYWxzZSwiaW5wdXRzIjpbeyJpbmRleGVkIjp0cnVlLCJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6Im93bmVyIiwidHlwZSI6ImFkZHJlc3MifSx7ImluZGV4ZWQiOnRydWUsImludGVybmFsVHlwZSI6ImFkZHJlc3MiLCJuYW1lIjoib3BlcmF0b3IiLCJ0eXBlIjoiYWRkcmVzcyJ9LHsiaW5kZXhlZCI6ZmFsc2UsImludGVybmFsVHlwZSI6ImJvb2wiLCJuYW1lIjoiYXBwcm92ZWQiLCJ0eXBlIjoiYm9vbCJ9XSwibmFtZSI6IkFwcHJvdmFsRm9yQWxsIiwidHlwZSI6ImV2ZW50In0seyJpbnB1dHMiOlt7ImludGVybmFsVHlwZSI6ImFkZHJlc3MiLCJuYW1lIjoidG8iLCJ0eXBlIjoiYWRkcmVzcyJ9LHsiaW50ZXJuYWxUeXBlIjoidWludDI1NiIsIm5hbWUiOiJ0b2tlbklkIiwidHlwZSI6InVpbnQyNTYifV0sIm5hbWUiOiJhcHByb3ZlIiwib3V0cHV0cyI6W10sInN0YXRlTXV0YWJpbGl0eSI6Im5vbnBheWFibGUiLCJ0eXBlIjoiZnVuY3Rpb24ifSx7ImlucHV0cyI6W10sIm5hbWUiOiJtaW50TkZUIiwib3V0cHV0cyI6W3siaW50ZXJuYWxUeXBlIjoidWludDI1NiIsIm5hbWUiOiIiLCJ0eXBlIjoidWludDI1NiJ9XSwic3RhdGVNdXRhYmlsaXR5Ijoibm9ucGF5YWJsZSIsInR5cGUiOiJmdW5jdGlvbiJ9LHsiYW5vbnltb3VzIjpmYWxzZSwiaW5wdXRzIjpbeyJpbmRleGVkIjp0cnVlLCJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6InByZXZpb3VzT3duZXIiLCJ0eXBlIjoiYWRkcmVzcyJ9LHsiaW5kZXhlZCI6dHJ1ZSwiaW50ZXJuYWxUeXBlIjoiYWRkcmVzcyIsIm5hbWUiOiJuZXdPd25lciIsInR5cGUiOiJhZGRyZXNzIn1dLCJuYW1lIjoiT3duZXJzaGlwVHJhbnNmZXJyZWQiLCJ0eXBlIjoiZXZlbnQifSx7ImlucHV0cyI6W10sIm5hbWUiOiJyZW5vdW5jZU93bmVyc2hpcCIsIm91dHB1dHMiOltdLCJzdGF0ZU11dGFiaWxpdHkiOiJub25wYXlhYmxlIiwidHlwZSI6ImZ1bmN0aW9uIn0seyJpbnB1dHMiOlt7ImludGVybmFsVHlwZSI6ImFkZHJlc3MiLCJuYW1lIjoiZnJvbSIsInR5cGUiOiJhZGRyZXNzIn0seyJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6InRvIiwidHlwZSI6ImFkZHJlc3MifSx7ImludGVybmFsVHlwZSI6InVpbnQyNTYiLCJuYW1lIjoidG9rZW5JZCIsInR5cGUiOiJ1aW50MjU2In1dLCJuYW1lIjoic2FmZVRyYW5zZmVyRnJvbSIsIm91dHB1dHMiOltdLCJzdGF0ZU11dGFiaWxpdHkiOiJub25wYXlhYmxlIiwidHlwZSI6ImZ1bmN0aW9uIn0seyJpbnB1dHMiOlt7ImludGVybmFsVHlwZSI6ImFkZHJlc3MiLCJuYW1lIjoiZnJvbSIsInR5cGUiOiJhZGRyZXNzIn0seyJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6InRvIiwidHlwZSI6ImFkZHJlc3MifSx7ImludGVybmFsVHlwZSI6InVpbnQyNTYiLCJuYW1lIjoidG9rZW5JZCIsInR5cGUiOiJ1aW50MjU2In0seyJpbnRlcm5hbFR5cGUiOiJieXRlcyIsIm5hbWUiOiJkYXRhIiwidHlwZSI6ImJ5dGVzIn1dLCJuYW1lIjoic2FmZVRyYW5zZmVyRnJvbSIsIm91dHB1dHMiOltdLCJzdGF0ZU11dGFiaWxpdHkiOiJub25wYXlhYmxlIiwidHlwZSI6ImZ1bmN0aW9uIn0seyJpbnB1dHMiOlt7ImludGVybmFsVHlwZSI6ImFkZHJlc3MiLCJuYW1lIjoib3BlcmF0b3IiLCJ0eXBlIjoiYWRkcmVzcyJ9LHsiaW50ZXJuYWxUeXBlIjoiYm9vbCIsIm5hbWUiOiJhcHByb3ZlZCIsInR5cGUiOiJib29sIn1dLCJuYW1lIjoic2V0QXBwcm92YWxGb3JBbGwiLCJvdXRwdXRzIjpbXSwic3RhdGVNdXRhYmlsaXR5Ijoibm9ucGF5YWJsZSIsInR5cGUiOiJmdW5jdGlvbiJ9LHsiaW5wdXRzIjpbeyJpbnRlcm5hbFR5cGUiOiJzdHJpbmciLCJuYW1lIjoiYmFzZVVSSSIsInR5cGUiOiJzdHJpbmcifV0sIm5hbWUiOiJzZXRCYXNlVVJJIiwib3V0cHV0cyI6W10sInN0YXRlTXV0YWJpbGl0eSI6Im5vbnBheWFibGUiLCJ0eXBlIjoiZnVuY3Rpb24ifSx7ImlucHV0cyI6W3siaW50ZXJuYWxUeXBlIjoidWludDI1NiIsIm5hbWUiOiJ0b3RhbFN1cHBseSIsInR5cGUiOiJ1aW50MjU2In1dLCJuYW1lIjoic2V0VG90YWxTdXBwbHkiLCJvdXRwdXRzIjpbXSwic3RhdGVNdXRhYmlsaXR5Ijoibm9ucGF5YWJsZSIsInR5cGUiOiJmdW5jdGlvbiJ9LHsiYW5vbnltb3VzIjpmYWxzZSwiaW5wdXRzIjpbeyJpbmRleGVkIjp0cnVlLCJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6ImZyb20iLCJ0eXBlIjoiYWRkcmVzcyJ9LHsiaW5kZXhlZCI6dHJ1ZSwiaW50ZXJuYWxUeXBlIjoiYWRkcmVzcyIsIm5hbWUiOiJ0byIsInR5cGUiOiJhZGRyZXNzIn0seyJpbmRleGVkIjp0cnVlLCJpbnRlcm5hbFR5cGUiOiJ1aW50MjU2IiwibmFtZSI6InRva2VuSWQiLCJ0eXBlIjoidWludDI1NiJ9XSwibmFtZSI6IlRyYW5zZmVyIiwidHlwZSI6ImV2ZW50In0seyJpbnB1dHMiOlt7ImludGVybmFsVHlwZSI6ImFkZHJlc3MiLCJuYW1lIjoiZnJvbSIsInR5cGUiOiJhZGRyZXNzIn0seyJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6InRvIiwidHlwZSI6ImFkZHJlc3MifSx7ImludGVybmFsVHlwZSI6InVpbnQyNTYiLCJuYW1lIjoidG9rZW5JZCIsInR5cGUiOiJ1aW50MjU2In1dLCJuYW1lIjoidHJhbnNmZXJGcm9tIiwib3V0cHV0cyI6W10sInN0YXRlTXV0YWJpbGl0eSI6Im5vbnBheWFibGUiLCJ0eXBlIjoiZnVuY3Rpb24ifSx7ImlucHV0cyI6W3siaW50ZXJuYWxUeXBlIjoiYWRkcmVzcyIsIm5hbWUiOiJuZXdPd25lciIsInR5cGUiOiJhZGRyZXNzIn1dLCJuYW1lIjoidHJhbnNmZXJPd25lcnNoaXAiLCJvdXRwdXRzIjpbXSwic3RhdGVNdXRhYmlsaXR5Ijoibm9ucGF5YWJsZSIsInR5cGUiOiJmdW5jdGlvbiJ9LHsiaW5wdXRzIjpbeyJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6Im93bmVyIiwidHlwZSI6ImFkZHJlc3MifV0sIm5hbWUiOiJiYWxhbmNlT2YiLCJvdXRwdXRzIjpbeyJpbnRlcm5hbFR5cGUiOiJ1aW50MjU2IiwibmFtZSI6IiIsInR5cGUiOiJ1aW50MjU2In1dLCJzdGF0ZU11dGFiaWxpdHkiOiJ2aWV3IiwidHlwZSI6ImZ1bmN0aW9uIn0seyJpbnB1dHMiOlt7ImludGVybmFsVHlwZSI6InVpbnQyNTYiLCJuYW1lIjoidG9rZW5JZCIsInR5cGUiOiJ1aW50MjU2In1dLCJuYW1lIjoiZ2V0QXBwcm92ZWQiLCJvdXRwdXRzIjpbeyJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6IiIsInR5cGUiOiJhZGRyZXNzIn1dLCJzdGF0ZU11dGFiaWxpdHkiOiJ2aWV3IiwidHlwZSI6ImZ1bmN0aW9uIn0seyJpbnB1dHMiOltdLCJuYW1lIjoiZ2V0VHJ1c3RlZEZvcndhcmRlciIsIm91dHB1dHMiOlt7ImludGVybmFsVHlwZSI6ImFkZHJlc3MiLCJuYW1lIjoiZm9yd2FyZGVyIiwidHlwZSI6ImFkZHJlc3MifV0sInN0YXRlTXV0YWJpbGl0eSI6InZpZXciLCJ0eXBlIjoiZnVuY3Rpb24ifSx7ImlucHV0cyI6W3siaW50ZXJuYWxUeXBlIjoiYWRkcmVzcyIsIm5hbWUiOiJvd25lciIsInR5cGUiOiJhZGRyZXNzIn0seyJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6Im9wZXJhdG9yIiwidHlwZSI6ImFkZHJlc3MifV0sIm5hbWUiOiJpc0FwcHJvdmVkRm9yQWxsIiwib3V0cHV0cyI6W3siaW50ZXJuYWxUeXBlIjoiYm9vbCIsIm5hbWUiOiIiLCJ0eXBlIjoiYm9vbCJ9XSwic3RhdGVNdXRhYmlsaXR5IjoidmlldyIsInR5cGUiOiJmdW5jdGlvbiJ9LHsiaW5wdXRzIjpbeyJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6ImZvcndhcmRlciIsInR5cGUiOiJhZGRyZXNzIn1dLCJuYW1lIjoiaXNUcnVzdGVkRm9yd2FyZGVyIiwib3V0cHV0cyI6W3siaW50ZXJuYWxUeXBlIjoiYm9vbCIsIm5hbWUiOiIiLCJ0eXBlIjoiYm9vbCJ9XSwic3RhdGVNdXRhYmlsaXR5IjoidmlldyIsInR5cGUiOiJmdW5jdGlvbiJ9LHsiaW5wdXRzIjpbXSwibmFtZSI6Im5hbWUiLCJvdXRwdXRzIjpbeyJpbnRlcm5hbFR5cGUiOiJzdHJpbmciLCJuYW1lIjoiIiwidHlwZSI6InN0cmluZyJ9XSwic3RhdGVNdXRhYmlsaXR5IjoidmlldyIsInR5cGUiOiJmdW5jdGlvbiJ9LHsiaW5wdXRzIjpbXSwibmFtZSI6Im93bmVyIiwib3V0cHV0cyI6W3siaW50ZXJuYWxUeXBlIjoiYWRkcmVzcyIsIm5hbWUiOiIiLCJ0eXBlIjoiYWRkcmVzcyJ9XSwic3RhdGVNdXRhYmlsaXR5IjoidmlldyIsInR5cGUiOiJmdW5jdGlvbiJ9LHsiaW5wdXRzIjpbeyJpbnRlcm5hbFR5cGUiOiJ1aW50MjU2IiwibmFtZSI6InRva2VuSWQiLCJ0eXBlIjoidWludDI1NiJ9XSwibmFtZSI6Im93bmVyT2YiLCJvdXRwdXRzIjpbeyJpbnRlcm5hbFR5cGUiOiJhZGRyZXNzIiwibmFtZSI6IiIsInR5cGUiOiJhZGRyZXNzIn1dLCJzdGF0ZU11dGFiaWxpdHkiOiJ2aWV3IiwidHlwZSI6ImZ1bmN0aW9uIn0seyJpbnB1dHMiOlt7ImludGVybmFsVHlwZSI6ImJ5dGVzNCIsIm5hbWUiOiJpbnRlcmZhY2VJZCIsInR5cGUiOiJieXRlczQifV0sIm5hbWUiOiJzdXBwb3J0c0ludGVyZmFjZSIsIm91dHB1dHMiOlt7ImludGVybmFsVHlwZSI6ImJvb2wiLCJuYW1lIjoiIiwidHlwZSI6ImJvb2wifV0sInN0YXRlTXV0YWJpbGl0eSI6InZpZXciLCJ0eXBlIjoiZnVuY3Rpb24ifSx7ImlucHV0cyI6W10sIm5hbWUiOiJzeW1ib2wiLCJvdXRwdXRzIjpbeyJpbnRlcm5hbFR5cGUiOiJzdHJpbmciLCJuYW1lIjoiIiwidHlwZSI6InN0cmluZyJ9XSwic3RhdGVNdXRhYmlsaXR5IjoidmlldyIsInR5cGUiOiJmdW5jdGlvbiJ9LHsiaW5wdXRzIjpbeyJpbnRlcm5hbFR5cGUiOiJ1aW50MjU2IiwibmFtZSI6InRva2VuSWQiLCJ0eXBlIjoidWludDI1NiJ9XSwibmFtZSI6InRva2VuVVJJIiwib3V0cHV0cyI6W3siaW50ZXJuYWxUeXBlIjoic3RyaW5nIiwibmFtZSI6IiIsInR5cGUiOiJzdHJpbmcifV0sInN0YXRlTXV0YWJpbGl0eSI6InZpZXciLCJ0eXBlIjoiZnVuY3Rpb24ifSx7ImlucHV0cyI6W10sIm5hbWUiOiJ2ZXJzaW9uUmVjaXBpZW50Iiwib3V0cHV0cyI6W3siaW50ZXJuYWxUeXBlIjoic3RyaW5nIiwibmFtZSI6IiIsInR5cGUiOiJzdHJpbmcifV0sInN0YXRlTXV0YWJpbGl0eSI6InZpZXciLCJ0eXBlIjoiZnVuY3Rpb24ifV0=",
    "method": "mintNFT",
    "params": []
}'
```
Remember to replace the placeholders with actual values as needed.