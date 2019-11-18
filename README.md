# tm-pkcs11

This is a PKCS#11 remote signer implementation for tendermint-based blockchain validators.

Limitations:

- Only support ECDSA-based validators (tendermint's default validator key is EdDSA)

## How to run hypermint and tm-pkcs11

```
$ ./tm-pkcs11
```

```
$ hmd start --log_level="main:info" --home=~/.hmd --priv_validator_laddr=tcp://0.0.0.0:26658
```

## How to use docker image

```
$ docker build . -t hypermint/tm-pkcs11:unstable
$ docker run -it --entrypoint /bin/bash hypermint/tm-pkcs11:unstable 
```

## References 

- https://aws.amazon.com/jp/blogs/security/how-to-run-aws-cloudhsm-workloads-on-docker-containers/
