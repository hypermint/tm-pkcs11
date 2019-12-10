package signerpv

import (
	"crypto/ecdsa"
	"crypto/rand"
	"github.com/btcsuite/btcd/btcec"
	"github.com/hypermint/tm-pkcs11/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	"os"
	"testing"
)

func TestGetPubKey(t *testing.T) {
	pv := createPV(t)
	assert.NotNil(t, pv)
	pubKey := pv.GetPubKey()
	assert.NotEmpty(t, pubKey.Address())
}

func TestSignMsg(t *testing.T) {
	a := assert.New(t)

	pv := createPV(t)
	msg := []byte{1, 2, 3}
	sigBytes, err := pv.signMsg(msg)
	a.NoError(err)
	a.NotNil(sigBytes)

	pub := pv.GetPubKey()
	a.True(pub.VerifyBytes(msg, sigBytes))
}

func createPV(t *testing.T) *SignerPV {
	options := []Option{
		RetryLimit(10),
		MalleableSigCheck(true),
	}

	_, found := os.LookupEnv("HSM_SOLIB")
	if found {
		createPVWithHSM(t, options...)
	}
	return createPVWithSigner(t, options...)
}

func createPVWithSigner(t *testing.T, options ...Option) *SignerPV {
	privKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		t.FailNow()
	}
	logger := log.NewTMLogger(
		log.NewSyncWriter(os.Stdout),
	).With("module", "signer")
	pv := NewSignerPV(privKey, btcec.S256(), logger, options...)
	if pv == nil {
		t.Fail()
	}
	return pv
}

func createPVWithHSM(t *testing.T, options ...Option) *SignerPV {
	maxSessions := 10
	id := helpers.RandomBytes32()

	solib, found := os.LookupEnv("HSM_SOLIB")
	if !found {
		solib = helpers.DefaultHsmSoLib
	}
	token, found := os.LookupEnv("HSM_TOKEN")
	if !found {
		token = "default"
	}
	password, found := os.LookupEnv("HSM_PASSWORD")
	if !found {
		password = "password"
	}
	c11, err := helpers.CreateCrypto11(solib, token, password, maxSessions)
	if err != nil {
		t.Fatal(err)
	}
	logger := log.NewTMLogger(
		log.NewSyncWriter(os.Stdout),
	).With("module", "signer")
	signer, err := c11.GenerateECDSAKeyPair(id, btcec.S256())
	if err != nil {
		t.Fatal(err)
	}
	pv := NewSignerPV(signer, btcec.S256(), logger, options...)
	if pv == nil {
		t.Fail()
	}
	return pv
}
