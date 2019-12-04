# Server-Side

## Setup

### mysql (e.g. port 8080)

```
$ mysql.server start
$ mysql -u root -p
mysql> create database acfts_8080
```

## Run (e.g. run with port 8080)

```
$ go run main.go
Input port number: 8080
...
[GIN-debug] Listening and serving HTTP on :8080/
```

## Example

```shell
#!/bin/bash

# Create genesis
curl -H 'Content-Type:application/json' -d "$JSON" http://localhost:8080/genesis

JSON=$(cat <<EOS
{
        "inputs": [
                {
                        "utxo": {
                                "address1": "example01",
                                "address2": "example02",
                                "previous_hash": "genesis",
                                "index": 0,
                                "server_signatures": [
                                        {
                                                "address1": "gene",
                                                "address2": "sis",
                                                "signature1": "dum",
                                                "signature2": "my"
                                        },
                                        {
                                                "address1": "gene",
                                                "address2": "sis",
                                                "signature1": "dum",
                                                "signature2": "my"
                                        }]
                        },
                        "siblings": [],
                        "signature1": "sig01",
                        "signature2": "sig02"
                }],
        "outputs": [
                {
                        "amount": 200,
                        "address1": "example11",
                        "address2": "example12",
                        "previous_hash": "a36d289b2ed18b196f07f59489710f45d92a97fdc2e2492dec67cf9cadb3a733",
                        "index": 0
                }]
}
EOS
)

curl -H 'Content-Type:application/json' -d "$JSON" http://localhost:8080/transaction
```

```
$ chmod 755 test.sh
$ ./test.sh
```

