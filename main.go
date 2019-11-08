package main

import (
	"crypto/elliptic"
	"flag"
	"fmt"
	"github.com/ThalesIgnite/crypto11"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	"math/rand"
	"os"
	"time"

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

	pv, err := CreateEcdsaPV(pkcs11lib)
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
		connTimeout := 3 * time.Second // TODO
		dialer = privval.DialTCPFn(address, connTimeout, ed25519.GenPrivKey())
		// dialer = xprivval.DialTCPFn(address, connTimeout, secp256k1.GenPrivKey())
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

func CreateEcdsaPV(pkcs11lib string) (types.PrivValidator, error) {
	if context, err := crypto11.Configure(&crypto11.Config{
		Path: pkcs11lib,
		TokenLabel: "hoge",
		Pin: "password",
		UseGCMIVFromHSM: true,
	}); err != nil {
		return nil, fmt.Errorf("failed to load PKCS#11 library err=%v path=%v", err, pkcs11lib)
	} else {
		id := randomBytes32()
		if signer, err := context.GenerateECDSAKeyPair(id, elliptic.P256()); err != nil {
			return nil, err
		} else {
			return NewRemoteSignerPV(signer), nil
		}
	}
}

func randomBytes32() []byte {
	result := make([]byte, 32)
	rand.Read(result)
	return result
}
