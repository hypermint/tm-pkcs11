# tm-pkcs11

This is a PKCS#11 remote signer implementation for tendermint-based blockchain validators.

Limitations:

- Only support ECDSA-based validators (tendermint's default validator key is EdDSA)

### Supported HSM

- SoftHSM v2
- AWS CloudHSM

This software just uses crypto11 library for signing votes.
Other PKCS#11-based HSM might work as well.

### Supported middlewares

Currently we tested with only hypermint (tendermint-based blockchain middleware with Web-assembly smart contract).
But this might works with other tendermint-based blockchains which use ECDSA key on their validator nodes. 

- hypermint (https://github.com/bluele/hypermint)

## How to run tm-pkcs11

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
addr = ":26658"
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

First, you have to build and run tm-pkcs11.
tm-pkcs11 will start trying to connect to the validator endpoint specified in the configuration.

```
$ ./tm-pkcs11
```

### Run hypermint

Run hmd command with priv_validator_laddr option and hypermint will accept the connection request from tm-pkcs11.

```
$ hmd start --log_level="main:info" --home=~/.hmd --priv_validator_laddr=tcp://0.0.0.0:26658
```

## Run with docker-compose

```
$ make build-image
$ make hm-config-from-image
$ make hm-init
$ docker-compose up
```

## How to use docker image

```
$ docker build . -t hypermint/tm-pkcs11:unstable
$ docker run -it hypermint/tm-pkcs11:unstable 
```
