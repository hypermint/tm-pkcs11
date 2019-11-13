package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/ThalesIgnite/crypto11"
	"github.com/hypermint/tm-pkcs11/remotepv"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	"os"
	"time"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	xprivval "github.com/hypermint/tm-pkcs11/privval"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
)

func main() {
	var (
		addr             = flag.String("addr", ":26656", "Address of client to connect to")
		chainID          = flag.String("chain-id", "test-chain-uAssCJ", "chain id")

		logger = log.NewTMLogger(
			log.NewSyncWriter(os.Stdout),
		).With("module", "priv_val")
	)
	flag.Parse()

	pkcs11lib, ok := os.LookupEnv("HSM_SOLIB")
	if !ok {
		logger.Error("HSM_SOLIB not set")
		os.Exit(1)
	}

	label := []byte("piyo3")
	c11ctx, err := remotepv.CreateCrypto11(pkcs11lib)
	if err != nil {
		panic(err)
	}
	if err := remotepv.GenerateKeyPair2(c11ctx, label); err != nil {
		logger.Info("failed to generate key pair", "error", err)
	}
	pv, err := CreateEcdsaPV(c11ctx, label)
	if err != nil {
		panic(err)
	}

	logger.Info(
		"Starting private validator",
		"addr", *addr,
		"chainID", *chainID,
	)

	var dialer privval.SocketDialer
	protocol, address := cmn.ProtocolAndAddress(*addr)
	switch protocol {
	case "unix":
		dialer = privval.DialUnixFn(address)
	case "tcp":
		connTimeout := 100 * time.Second // TODO
		dialer = xprivval.DialTCPFn(address, connTimeout, secp256k1.GenPrivKey())
	default:
		logger.Error("Unknown protocol", "protocol", protocol)
		os.Exit(1)
	}

	sd := privval.NewSignerDialerEndpoint(logger, dialer)
	ss := privval.NewSignerServer(sd, *chainID, pv)

	if err := ss.Start(); err != nil {
		panic(err)
	}

	// Stop upon receiving SIGTERM or CTRL-C.
	cmn.TrapSignal(logger, func() {
		err := ss.Stop()
		if err != nil {
			panic(err)
		}
	})

	// Run forever.
	select {}
}

func CreateEcdsaPV(context *crypto11.Context, label []byte) (types.PrivValidator, error) {
	if signer, err := context.FindKeyPair(nil, label); err != nil {
		return nil, err
	} else {
		logger := log.NewTMLogger(
			log.NewSyncWriter(os.Stdout),
		).With("module", "signer")
		if signer == nil {
			return nil, fmt.Errorf("signer is nil")
		}
		pubKey0, ok := signer.Public().(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not a ECDSA key")
		}

		pubKey, err := remotepv.PublicKeyToPubKeySecp256k1(pubKey0)
		if err != nil {
			return nil, err
		}
		logger.Info("validator key info",
			"address", pubKey.Address(),
			"address_eth", gethcrypto.PubkeyToAddress(*pubKey0).Hex(),
			"pub_key", cdc.MustMarshalJSON(pubKey),
			"pub_key_bytes", hex.EncodeToString(pubKey[:]),
		)
		return remotepv.NewRemoteSignerPV(signer, logger), nil
	}
}
