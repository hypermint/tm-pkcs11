package helpers

import (
	"errors"
	"fmt"
	gincocrypto "github.com/GincoInc/go-crypto"
	"github.com/ThalesIgnite/crypto11"
	"github.com/miekg/pkcs11"
)

const DefaultHsmSoLib = "/usr/local/lib/softhsm/libsofthsm2.so"

var (
	ErrKeyFound = errors.New("key found")
)

func CreateCrypto11(pkcs11lib, tokenLabel, password string) (*crypto11.Context, error) {
	context, err := crypto11.Configure(&crypto11.Config{
		Path: pkcs11lib,
		TokenLabel: tokenLabel,
		Pin: password,
		UseGCMIVFromHSM: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load PKCS#11 library err=%v path=%v", err, pkcs11lib)
	}
	return context, nil
}

func GenerateKeyPair(context *crypto11.Context, label []byte) error {
	id := RandomBytes32()
	_, err := context.GenerateECDSAKeyPairWithLabel(id, label, gincocrypto.Secp256k1())
	if err != nil {
		return err
	}
	return nil
}

func GenerateKeyPair2(context *crypto11.Context, label []byte) error {
	pubId := RandomBytes32()
	if pub, err := crypto11.NewAttributeSetWithIDAndLabel(pubId, label); err != nil {
		return err
	} else {
		priv := pub.Copy()
		pub.AddIfNotPresent([]*pkcs11.Attribute{
			// https://www.secg.org/sec2-v2.pdf
			// https://play.golang.org/p/M0VLD0RZAaM
			pkcs11.NewAttribute(pkcs11.CKA_ECDSA_PARAMS, []byte {0x06, 0x05, 0x2b, 0x81, 0x04, 0x00, 0x0a}),
		})
		if signer, err := context.GenerateECDSAKeyPairWithAttributes(pub, priv, gincocrypto.Secp256k1()); err != nil {
			return err
		} else {
			if signer == nil {
				return fmt.Errorf("signer is nil")
			}
			return nil
		}
	}
}
