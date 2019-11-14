
MNEMONIC="token dash time stand brisk fatal health honey frozen brown flight kitchen"
HDW_PATH="m/44'/60'/0'/0/0"
HSM_SOLIB=/usr/local/lib/softhsm/libsofthsm2.so # for homebrew

TM_PARAMS='log_level="*:error"'
export TM_PARAMS

build:
	go build

build-image:
	docker build . -t hypermint/tm-pkcs11:unstable

run-image:
	docker run -it hypermint/tm-pkcs11:unstable

hm-init:
	rm -rf /tmp/hypermint
	docker run -it --rm -v "/tmp/hypermint:/root/.hmd" bluele/hypermint:unstable /hmd tendermint init-validator --mnemonic $(MNEMONIC) --hdw_path $(HDW_PATH)
	docker run -it --rm -v "/tmp/hypermint:/root/.hmd" -v "$(PWD)/hm-config:/hm-config" bluele/hypermint:unstable /hmd create --genesis /hm-config/genesis.json

run-tm-pkcs11:
	HSM_SOLIB=$(HSM_SOLIB) ./tm-pkcs11 --addr 127.0.0.1:26658 --chain-id test-chain-uAssCJ

validate-circleci-config:
	circleci config validate .circleci/config.yml
