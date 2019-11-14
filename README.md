# tm-pkcs11

This is a PKCS#11 remote signer implementation for tendermint-based blockchain validators.

Limitations:

- Only support ECDSA-based validators (tendermint's default validator key is EdDSA)

```
./build/hmd start --log_level="main:info" --home=~/.hmd --priv_validator_laddr=tcp://0.0.0.0:26658
```

```
$ ./tm-pkcs11 --addr :26658 --priv-key ~/.hmd/config/priv_validator_key.json
```

## How to use docker image 

```
$ docker run -it --entrypoint /bin/bash hypermint/tm-pkcs11:unstable 
```

## References 

- https://aws.amazon.com/jp/blogs/security/how-to-run-aws-cloudhsm-workloads-on-docker-containers/
