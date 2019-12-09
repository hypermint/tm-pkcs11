module github.com/hypermint/tm-pkcs11

go 1.13

require (
	github.com/ThalesIgnite/crypto11 v1.2.1
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/ethereum/go-ethereum v1.8.21
	github.com/libp2p/go-buffer-pool v0.0.2
	github.com/miekg/pkcs11 v1.0.3
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.1
	github.com/spf13/viper v1.5.0
	github.com/stretchr/testify v1.4.0
	github.com/tendermint/go-amino v0.14.1
	github.com/tendermint/tendermint v0.32.8
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
)

replace github.com/ThalesIgnite/crypto11 => github.com/hypermint/crypto11 v1.2.2-0.20191206031436-3b6bf5a91977
