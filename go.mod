module github.com/hypermint/tm-pkcs11

go 1.13

require (
	github.com/GincoInc/go-crypto v1.2.0
	github.com/ThalesIgnite/crypto11 v1.2.1
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/btcsuite/btcd v0.0.0-20190115013929-ed77733ec07d
	github.com/ethereum/go-ethereum v1.8.21
	github.com/libp2p/go-buffer-pool v0.0.2
	github.com/miekg/pkcs11 v1.0.3
	github.com/pkg/errors v0.8.1
	github.com/tendermint/go-amino v0.14.1
	github.com/tendermint/tendermint v0.32.3
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
)

replace github.com/ThalesIgnite/crypto11 => ../crypto11
