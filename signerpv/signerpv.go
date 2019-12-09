package signerpv

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
	"math/big"
)

var errRetryLimitExceeded = errors.New("exceeded retry limit")

var secp256k1halfN = new(big.Int).Rsh(btcec.S256().N, 1)

type SignerPV struct {
	s                 crypto.Signer
	curve             elliptic.Curve
	logger            log.Logger
	retryLimit        int
	malleableSigCheck bool

	pubKeyCache tmcrypto.PubKey
	retry       int64
}

func NewSignerPV(signer crypto.Signer, curve elliptic.Curve, logger log.Logger, options... Option) *SignerPV {
	pv := &SignerPV{
		s:                 signer,
		curve:             curve,
		logger:            logger,
		retryLimit:        100,
		malleableSigCheck: true,
	}

	for _, option := range options {
		option(pv)
	}

	return pv
}

func (pv *SignerPV) GetPubKey() tmcrypto.PubKey {
	if pv.pubKeyCache != nil {
		return pv.pubKeyCache
	}
	signer := pv.s
	pk0 := signer.Public()
	if pk, ok := pk0.(*ecdsa.PublicKey); ok {
		p := PublicKeyToPubKeySecp256k1(pk)
		pv.logger.Debug("GetPubKey", "address", p.Address(), "pubkey", hex.EncodeToString(p[:]))
		pv.pubKeyCache = p
		return p
	} else {
		return nil
	}
}

func (pv *SignerPV) SignVote(chainID string, vote *types.Vote) error {
	if sigBytes, err := pv.signMsg(vote.SignBytes(chainID)); err != nil {
		return err
	} else {
		pv.logger.Debug("SignVote", "chainID", chainID, "sig", hex.EncodeToString(sigBytes), "sig_len", len(sigBytes))
		vote.Signature = sigBytes
		return nil
	}
}

func (pv *SignerPV) SignProposal(chainID string, proposal *types.Proposal) error {
	if sigBytes, err := pv.signMsg(proposal.SignBytes(chainID)); err != nil {
		return err
	} else {
		pv.logger.Debug("SignProposal", "chainID", chainID, "sig", hex.EncodeToString(sigBytes), "sig_len", len(sigBytes))
		proposal.Signature = sigBytes
		return nil
	}
}

func (pv *SignerPV) signMsg(msgBytes []byte) ([]byte, error) {
	hash := tmcrypto.Sha256(msgBytes)
	var sigBytes []byte

	retryCount := 0
	for sigBytes == nil && retryCount < pv.retryLimit {
		derSig, err := pv.s.Sign(rand.Reader, hash[:], nil)
		if err != nil {
			return nil, err
		}

		signature, err := btcec.ParseDERSignature(derSig, pv.curve)
		if err != nil {
			return nil, err
		}

		// Reject malleable signatures.
		// Check tendermint@v0.32.3/crypto/secp256k1/secp256k1_nocgo.go for detail
		if pv.malleableSigCheck && signature.S.Cmp(secp256k1halfN) > 0 {
			retryCount++
			continue
		}

		sigBytes = make([]byte, 64)
		rbytes, sbytes := signature.R.Bytes(), signature.S.Bytes()
		copy(sigBytes[32-len(rbytes):32], rbytes)
		copy(sigBytes[64-len(sbytes):64], sbytes)
	}

	pv.retry += int64(retryCount)
	if sigBytes == nil {
		return nil, errRetryLimitExceeded
	}
	return sigBytes, nil
}

func PublicKeyToPubKeySecp256k1(pubKey0 *ecdsa.PublicKey) secp256k1.PubKeySecp256k1 {
	pubKey := btcec.PublicKey(*pubKey0)
	var tmPubkeyBytes secp256k1.PubKeySecp256k1
	copy(tmPubkeyBytes[:], pubKey.SerializeCompressed())
	return tmPubkeyBytes
}
