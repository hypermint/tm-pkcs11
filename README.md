# tm-pkcs11

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
