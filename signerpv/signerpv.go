package signerpv

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
	"math/big"
)

var secp256k1halfN = new(big.Int).Rsh(btcec.S256().N, 1)

type SignerPV struct {
	s crypto.Signer
	curve elliptic.Curve
	logger log.Logger

	pubKeyCache tmcrypto.PubKey
}

func NewSignerPV(signer crypto.Signer, curve elliptic.Curve, logger log.Logger) *SignerPV {
	return &SignerPV{
		signer,
		curve,
		logger,
		nil,
	}
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
	if derSig, err := pv.s.Sign(rand.Reader, hash[:], nil); err != nil {
		return nil, err
	} else {
		signature, err := btcec.ParseDERSignature(derSig, pv.curve)
		if err != nil {
			return nil, err
		}

		rbytes, sbytes := signature.R.Bytes(), signature.S.Bytes()
		sigBytes := make([]byte, 64)
		copy(sigBytes[32-len(rbytes):32], rbytes)
		copy(sigBytes[64-len(sbytes):64], sbytes)

		// Reject malleable signatures.
		// Check tendermint@v0.32.3/crypto/secp256k1/secp256k1_nocgo.go for detail
		if signature.S.Cmp(secp256k1halfN) > 0 {
			return pv.signMsg(msgBytes)
		}

		return sigBytes, nil
	}
}

func PublicKeyToPubKeySecp256k1(pubKey0 *ecdsa.PublicKey) secp256k1.PubKeySecp256k1 {
	pubKey := btcec.PublicKey(*pubKey0)
	var tmPubkeyBytes secp256k1.PubKeySecp256k1
	copy(tmPubkeyBytes[:], pubKey.SerializeCompressed())
	return tmPubkeyBytes
}
