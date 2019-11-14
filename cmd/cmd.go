package cmd

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ThalesIgnite/crypto11"
	"github.com/hypermint/tm-pkcs11/helpers"
	"github.com/hypermint/tm-pkcs11/remotepv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	"os"
	"strings"
	"time"

	xprivval "github.com/hypermint/tm-pkcs11/privval"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	rootCmd.PersistentFlags().String("addr", ":26656", "Address of client to connect to")
	rootCmd.PersistentFlags().String("chain-id", "test-chain-uAssCJ", "chain id")
	rootCmd.PersistentFlags().String("key-label", "default", "key label")
	rootCmd.PersistentFlags().String("token-label", "hoge", "token label")
	rootCmd.PersistentFlags().String("password", "password", "password")
}

var rootCmd = &cobra.Command{
	Use:   "tm-pkcs11",
	Short: "PKCS#11 remote signer for tendermint-based validator",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}

		addr := viper.GetString("addr")
		chainID := viper.GetString("chain-id")
		keyLabel := viper.GetString("key-label")
		tokenLabel := viper.GetString("token-label")
		password := viper.GetString("password")

		logger := log.NewTMLogger(
			log.NewSyncWriter(os.Stdout),
		).With("module", "priv_val")

		pkcs11lib, ok := os.LookupEnv("HSM_SOLIB")
		if !ok {
			pkcs11lib = helpers.DefaultHsmSoLib
		}

		c11ctx, err := helpers.CreateCrypto11(pkcs11lib, tokenLabel, password)
		if err != nil {
			panic(err)
		}

		if err := helpers.GenerateKeyPair2(c11ctx, []byte(keyLabel)); err != nil {
			if err != helpers.ErrKeyFound {
				panic(err)
			}
		}
		pv, err := CreateEcdsaPV(c11ctx, []byte(keyLabel))
		if err != nil {
			panic(err)
		}

		logger.Info("Starting private validator", "addr", addr, "chainID", chainID, )

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
		ss := privval.NewSignerServer(sd, chainID, pv)

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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v", err.Error())
		os.Exit(1)
	}
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
			"pub_key", cdc.MustMarshalJSON(pubKey),
		)
		return remotepv.NewRemoteSignerPV(signer, logger), nil
	}
}
