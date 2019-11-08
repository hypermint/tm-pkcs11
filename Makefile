
MNEMONIC="token dash time stand brisk fatal health honey frozen brown flight kitchen"
HDW_PATH="m/44'/60'/0'/0/0"

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

tm-init:
	rm -rf /tmp/tendermint
	docker run -it --rm -v "/tmp/tendermint:/tendermint" -v "$(PWD)/tm-config:/tm-config" tendermint/tendermint init
	cp tm-config/* /tmp/tendermint/config/
