package cmd

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ThalesIgnite/crypto11"
	"github.com/btcsuite/btcd/btcec"
	"github.com/hypermint/tm-pkcs11/helpers"
	"github.com/hypermint/tm-pkcs11/signerpv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	"os"
	"time"

	xprivval "github.com/hypermint/tm-pkcs11/privval"
	cmn "github.com/tendermint/tendermint/libs/common"
)

const (
	FlagAddr                   = "addr"
	FlagChainId                = "chain-id"
	FlagKeyLabel               = "key-label"
	FlagTokenLabel             = "token-label"
	FlagPassword               = "password"
	FlagHsmSolib               = "hsm-solib"
	FlagDialerConnRetries      = "dialer-conn-retries"
	FlagDialerReadWriteTimeout = "dialer-rw-timeout"
	FlagLogLevel               = "log-level"
	FlagMaxSessions            = "max-sessions"
)

const (
	DefaultKeyLabel = "default"
	DefaultTokenLabel = "default"
	DefaultPassword = "password"
	DefaultLogLevel = "info"
	DefaultMaxSessions = 512
	DefaultDialerConnRetries = 1000
	DefaultDialerReadWriteTimeout = 3000
)

func init() {
	rootCmd.Flags().String(FlagAddr, ":26658", "Address of client to connect to")
	rootCmd.Flags().String(FlagChainId, "test-chain", "chain id")
	rootCmd.Flags().String(FlagKeyLabel, DefaultKeyLabel, "key label")
	rootCmd.Flags().Int(FlagDialerConnRetries, DefaultDialerConnRetries, "retry limit of dialer")
	rootCmd.Flags().Int(FlagDialerReadWriteTimeout, DefaultDialerReadWriteTimeout, "read/write timeout in millisecond")
	rootCmd.Flags().Int(FlagMaxSessions, DefaultMaxSessions, "max sessions")
	rootCmd.PersistentFlags().String(FlagTokenLabel, DefaultTokenLabel, "token label")
	rootCmd.PersistentFlags().String(FlagPassword, DefaultPassword, "password")
	rootCmd.PersistentFlags().String(FlagHsmSolib, helpers.DefaultHsmSoLib, "password")
	rootCmd.PersistentFlags().String(FlagLogLevel, DefaultLogLevel, "log level")
}

var rootCmd = &cobra.Command{
	Use:   "tm-pkcs11",
	Short: "PKCS#11 remote signer for tendermint-based validator",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		addr := viper.GetString(FlagAddr)
		chainID := viper.GetString(FlagChainId)
		keyLabel := viper.GetString(FlagKeyLabel)
		tokenLabel := viper.GetString(FlagTokenLabel)
		password := viper.GetString(FlagPassword)
		hsmSolib := viper.GetString(FlagHsmSolib)
		dialerConnRetries := viper.GetInt(FlagDialerConnRetries)
		dialerTimeoutReadWrite := viper.GetInt(FlagDialerReadWriteTimeout)
		maxSessions := viper.GetInt(FlagMaxSessions)

		logger := log.NewTMLogger(
			log.NewSyncWriter(os.Stdout),
		).With("module", "priv_val")

		logLevel := viper.GetString(FlagLogLevel)
		if opt, err := log.AllowLevel(logLevel); err != nil {
			return err
		} else {
			logger = log.NewFilter(logger, opt)
		}

		c11ctx, err := helpers.CreateCrypto11(hsmSolib, tokenLabel, password, maxSessions)
		if err != nil {
			panic(err)
		}

		pv, err := CreateEcdsaPV(c11ctx, []byte(keyLabel))
		if err != nil {
			panic(err)
		}

		logger.Info("Starting private validator", "addr", addr, "chainID", chainID)

		var dialer privval.SocketDialer
		protocol, address := cmn.ProtocolAndAddress(addr)
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
		privval.SignerDialerEndpointConnRetries(dialerConnRetries)(sd)
		privval.SignerDialerEndpointTimeoutReadWrite(time.Duration(dialerTimeoutReadWrite)*time.Millisecond)(sd)
		ss := privval.NewSignerServer(sd, chainID, pv)
		ss.SetLogger(logger)

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
	},
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

		pubKey := signerpv.PublicKeyToPubKeySecp256k1(pubKey0)
		logger.Info("validator key info",
			"address", pubKey.Address(),
			"pub_key", cdc.MustMarshalJSON(pubKey),
		)
		return signerpv.NewSignerPV(signer, btcec.S256(), logger), nil
	}
}
