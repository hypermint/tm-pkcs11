package remotepv

import (
	"fmt"
	gincocrypto "github.com/GincoInc/go-crypto"
	"github.com/ThalesIgnite/crypto11"
	"github.com/miekg/pkcs11"
	"math/rand"
)

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

func randomBytes32() []byte {
	result := make([]byte, 32)
	rand.Read(result)
	return result
}
