package main

import (
	"flag"
	"github.com/tendermint/tendermint/privval"
	"os"
	"time"

	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"

	xprivval "github.com/datachainlab/tm-pkcs11/privval"
)

func main() {
	var (
		addr             = flag.String("addr", ":26656", "Address of client to connect to")
		chainID          = flag.String("chain-id", "mychain", "chain id")
		privValKeyPath   = flag.String("priv-key", "", "priv val key file path")
		privValStatePath = flag.String("priv-state", "", "priv val state file path")

		logger = log.NewTMLogger(
			log.NewSyncWriter(os.Stdout),
		).With("module", "priv_val")
	)
	flag.Parse()

	logger.Info(
		"Starting private validator",
		"addr", *addr,
		"chainID", *chainID,
		"privKeyPath", *privValKeyPath,
		"privStatePath", *privValStatePath,
	)

	pv := privval.LoadFilePVEmptyState(*privValKeyPath, *privValStatePath)

	var dialer privval.SocketDialer
	protocol, address := cmn.ProtocolAndAddress(*addr)
	switch protocol {
	case "unix":
		dialer = xprivval.DialUnixFn(address)
	case "tcp":
		connTimeout := 3 * time.Second // TODO
		dialer = xprivval.DialTCPFn(address, connTimeout, ed25519.GenPrivKey())
	default:
		logger.Error("Unknown protocol", "protocol", protocol)
		os.Exit(1)
	}

	sd := privval.NewSignerDialerEndpoint(logger, dialer)
	ss := privval.NewSignerServer(sd, *chainID, pv)

	err := ss.Start()
	if err != nil {
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
