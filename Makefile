
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

inspect-image:
	docker run -it hypermint/tm-pkcs11:unstable pubkey

hm-config-from-image:
	pub_key_value=$$(docker run hypermint/tm-pkcs11:unstable pubkey | jq -r .value); \
	address=$$(docker run hypermint/tm-pkcs11:unstable pubkey --show-address); \
	tmpfile=$$(mktemp); \
	cat hm-config/genesis.json | jq ".validators[].pub_key.value = \"$${pub_key_value}\" | .validators[].address = \"$${address}\"" > $${tmpfile}; \
	cp $${tmpfile} hm-config/genesis.json

hm-init:
	rm -rf /tmp/hypermint
	docker run --rm -v "/tmp/hypermint:/root/.hmd" bluele/hypermint:unstable /hmd tendermint init-validator --mnemonic $(MNEMONIC) --hdw_path $(HDW_PATH)
	docker run --rm -v "/tmp/hypermint:/root/.hmd" -v "$(PWD)/hm-config:/hm-config" bluele/hypermint:unstable /hmd create --genesis /hm-config/genesis.json

run-tm-pkcs11:
	HSM_SOLIB=$(HSM_SOLIB) ./tm-pkcs11 --addr 127.0.0.1:26658 --chain-id test-chain-uAssCJ --token-label default --log-level debug

validate-circleci-config:
	circleci config validate .circleci/config.yml
