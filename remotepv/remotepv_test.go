package remotepv

import (
	gincocrypto "github.com/GincoInc/go-crypto"
	"github.com/hypermint/tm-pkcs11/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	"os"
	"testing"
)

func TestGetPubKey(t *testing.T) {
	id := helpers.RandomBytes32()
	pv := createPV(id, t)
	assert.NotNil(t, pv)
	pubKey := pv.GetPubKey()
	assert.NotEmpty(t, pubKey.Address())
}

func TestSignMsg(t *testing.T) {
	a := assert.New(t)
	id := helpers.RandomBytes32()
	pv := createPV(id, t)
	msg := []byte{1, 2, 3}
	sig, err := pv.signMsg(msg)
	a.NoError(err)
	a.NotNil(sig)
}

func createPV(id []byte, t *testing.T) *RemoteSignerPV {
	solib, found := os.LookupEnv("HSM_SOLIB")
	if !found {
		solib = helpers.DefaultHsmSoLib
	}
	c11, err := helpers.CreateCrypto11(solib, "hoge", "password")
	if err != nil {
		t.Fatal(err)
	}
	logger := log.NewTMLogger(
		log.NewSyncWriter(os.Stdout),
	).With("module", "signer")
	signer, err := c11.GenerateECDSAKeyPair(id, gincocrypto.Secp256k1())
	if err != nil {
		t.Fatal(err)
	}
	pv := NewRemoteSignerPV(signer, logger)
	if pv == nil {
		t.Fail()
	}
	return pv
}
