
build:
	go build

build-image:
	docker build . -t hypermint/tm-pkcs11:unstable

run-image:
	docker run hypermint/tm-pkcs11:unstable
