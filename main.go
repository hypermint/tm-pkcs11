package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/ThalesIgnite/crypto11"
	"github.com/miekg/pkcs11"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	"math/rand"
	"os"
	"time"

	gincocrypto "github.com/GincoInc/go-crypto"
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
	c11ctx, err := CreateCrypto11(pkcs11lib)
	if err != nil {
		panic(err)
	}
	if err := GenerateKeyPair2(c11ctx, label); err != nil {
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

func CreateCrypto11(pkcs11lib string) (*crypto11.Context, error) {
	context, err := crypto11.Configure(&crypto11.Config{
		Path: pkcs11lib,
		TokenLabel: "hoge",
		Pin: "password",
		UseGCMIVFromHSM: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load PKCS#11 library err=%v path=%v", err, pkcs11lib)
	}
	return context, nil
}

func GenerateKeyPair(context *crypto11.Context, label []byte) error {
	id := randomBytes32()
	if _, err := context.FindKeyPair(nil, label); err == nil {
		return fmt.Errorf("key found: %v", label)
	}
	_, err := context.GenerateECDSAKeyPairWithLabel(id, label, gincocrypto.Secp256k1())
	if err != nil {
		return err
	}
	return nil
}

func GenerateKeyPair2(context *crypto11.Context, label []byte) error {
	if signer, err := context.FindKeyPair(nil, label); err != nil {
		return err
	} else if signer != nil {
		return fmt.Errorf("key found: %v", label)
	}
	pubId := randomBytes32()
	dummyCurve := gincocrypto.Secp256k1()
	if pub, err := crypto11.NewAttributeSetWithIDAndLabel(pubId, label); err != nil {
		return err
	} else {
		priv := pub.Copy()
		pub.AddIfNotPresent([]*pkcs11.Attribute{
			// https://www.secg.org/sec2-v2.pdf
			// https://play.golang.org/p/M0VLD0RZAaM
			pkcs11.NewAttribute(pkcs11.CKA_ECDSA_PARAMS, []byte {0x06, 0x05, 0x2b, 0x81, 0x04, 0x00, 0x0a}),
		})
		if signer, err := context.GenerateECDSAKeyPairWithAttributes(pub, priv, dummyCurve); err != nil {
			return err
		} else {
			if signer == nil {
				return fmt.Errorf("signer is nil")
			}
			return nil
		}
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
		// address := gethcrypto.PubkeyToAddress(*pubKey0)
		pubKey := secp256k1.PubKeySecp256k1{}
		pubKeyBytes := gethcrypto.CompressPubkey(pubKey0)
		if len(pubKeyBytes) != secp256k1.PubKeySecp256k1Size {
			return nil, fmt.Errorf("invalid pubkey length: %v", len(pubKeyBytes))
		}
		copy(pubKey[:], pubKeyBytes[1:])
		pubKey[secp256k1.PubKeySecp256k1Size-1] = 0
		logger.Info("validator",
			"address", pubKey.Address(),
			"pub_key", cdc.MustMarshalJSON(pubKey),
			"pub_key_bytes", hex.EncodeToString(pubKey[:]),
		)
		return NewRemoteSignerPV(signer, logger), nil
	}
}

func randomBytes32() []byte {
	result := make([]byte, 32)
	rand.Read(result)
	return result
}
