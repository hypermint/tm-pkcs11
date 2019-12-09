# tm-pkcs11

This is a PKCS#11 remote signer implementation for tendermint-based blockchain validators.
(Please note that this software is still under development and not ready for production use)

Limitations:
- Only support ECDSA-based validators (tendermint's default validator key is EdDSA)

TODO:
- Vault support (e.g. AWS KSM) for storing HSM passwords

### Supported HSM

- SoftHSM v2
- AWS CloudHSM

This software just uses crypto11 library for signing votes.
Other PKCS#11-based HSM might work as well.

### Supported middlewares

Currently we test this software with "hypermint" (tendermint-based blockchain middleware with WebAssembly smart contract).
But this might works with other tendermint-based blockchains which use ECDSA key for their validator kes. 

- hypermint (https://github.com/bluele/hypermint)

## How to run tm-pkcs11 with SoftHSM

### Setting up SoftHSM

Before you use tm-pkcs11, you need to setup HSM and import or generate a EC keypair on HSM.

For example, you can initialize a token with the following command. 

```
$ softhsm2-util --init-token --slot 0 --label "default" --so-pin password --pin password
```

### Configure tm-pkcs11 and generate EC key

Then, put your own configuration in config.toml.

```
hsm-solib = "/usr/local/lib/softhsm/libsofthsm2.so" # for Mac
token-label = "default"
chain-id = "test-chain-uAssCJ" # please replace it with your chain id 
addr = ":26658" # signer endpoint
```

Now, you can generate a EC keypair using the following command.

```
$ cd tm-pkcs11
$ go build
$ ./tm-pkcs11 genkey --key-label default
{"type":"tendermint/PubKeySecp256k1","value":"AsNsIaQdj2ov4RGMZFIF6wGpCFWP714pTkVdqcGGG/bE"}
```

Check the address corresponding to the public key.

```
$ ./tm-pkcs11 pubkey --key-label default --show-address
3E5C835FDBF92AB2BD44A82D6B8ED99C6841DD7D
```

### Configure hypermint

You can put the generated key into your genesis.json file.

```
  "validators": [
    {
      "address": "3E5C835FDBF92AB2BD44A82D6B8ED99C6841DD7D",
      "pub_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "AsNsIaQdj2ov4RGMZFIF6wGpCFWP714pTkVdqcGGG/bE"
      },
      "power": "1",
      "name": ""
    }
  ],
```

### Run tm-pkcs11

First, you have to run tm-pkcs11.
tm-pkcs11 will start trying to connect to the signer endpoint specified in the configuration.

```
$ ./tm-pkcs11
```

### Run hypermint

Next, open another terminal window and run "hmd" command with priv_validator_laddr option.
hypermint will immediately accept the connection request from tm-pkcs11 and start working.

```
$ hmd start --log_level="main:info" --home=~/.hmd --priv_validator_laddr=tcp://0.0.0.0:26658
```

## Run hypermint and tm-pkcs11 using docker-compose

```
$ make build-image
$ make hm-config-from-image
$ make hm-init
$ docker-compose up
```

## How to use docker image

Docker image includes an entrypoint script which runs required daemons before running
tm-pkcs11. This image uses SoftHSM by default for testing purposes.

```
$ docker build . -t hypermint/tm-pkcs11:unstable
$ docker run -it hypermint/tm-pkcs11:unstable 
```

Please set HSM environment variable for other HSMs and specify your configurations as well.
For CloudHSM, it needs ENI IP address of a HSM module. 

```
$ docker run -it \
  -e HSM=cloudhsm \
  -e CLOUDHSM_IP=10.101.9.111 \
  hypermint/tm-pkcs11:unstable \
  --addr :26658 \
  --token-label cavium \
  --key-label validator1 \
  --password val1:password
```

(Some variables should not be included in an image nor command-line parameters.)
