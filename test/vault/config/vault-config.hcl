storage "file" {
  path = "vault/data"
}
listener "tcp" {
  address  = "0.0.0.0:8200"
  tls_disable = 1
}
seal "awskms" {
  region     = "ap-south-1"
  access_key = "AKIASLFDSYOEMZZG3HFL"
  secret_key = "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"
  kms_key_id = "ed95a2ea-bb69-4369-9969-0ca63854f41b"
}
ui = true
api_addr     = "http://127.0.0.1:8200"
