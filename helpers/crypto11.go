package helpers

import (
	"errors"
	"fmt"
	"github.com/ThalesIgnite/crypto11"
	"github.com/btcsuite/btcd/btcec"
	"github.com/miekg/pkcs11"
)

const DefaultHsmSoLib = "/usr/local/lib/softhsm/libsofthsm2.so"

type PKCryptoType string

const (
	Secp256k1 PKCryptoType = "ecdsa.secp256k1"
)

var (
	ErrKeyFound = errors.New("key found")
)

func CreateCrypto11(pkcs11lib, tokenLabel, password string, maxSessions int) (*crypto11.Context, error) {
	context, err := crypto11.Configure(&crypto11.Config{
		Path: pkcs11lib,
		TokenLabel: tokenLabel,
		Pin: password,
		UseGCMIVFromHSM: true,
		MaxSessions: maxSessions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load PKCS#11 library err=%v path=%v", err, pkcs11lib)
	}
	return context, nil
}

func GenerateKeyPair(context *crypto11.Context, label []byte, ecType PKCryptoType) error {
	pubId := RandomBytes32()
	if pub, err := crypto11.NewAttributeSetWithIDAndLabel(pubId, label); err != nil {
		return err
	} else {
		priv := pub.Copy()
		switch ecType {
		case Secp256k1:
			pub.AddIfNotPresent([]*pkcs11.Attribute{
				// https://www.secg.org/sec2-v2.pdf
				// https://play.golang.org/p/M0VLD0RZAaM
				pkcs11.NewAttribute(pkcs11.CKA_ECDSA_PARAMS, []byte {0x06, 0x05, 0x2b, 0x81, 0x04, 0x00, 0x0a}),
			})
			if signer, err := context.GenerateECDSAKeyPairWithAttributes(pub, priv, btcec.S256()); err != nil {
				return err
			} else {
				if signer == nil {
					return fmt.Errorf("signer is nil")
				}
				return nil
			}
		default:
			return errors.New("unsupported EC type")
		}
	}
}
