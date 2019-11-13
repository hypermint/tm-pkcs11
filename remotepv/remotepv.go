package remotepv

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	gincocrypto "github.com/GincoInc/go-crypto"
	"github.com/ThalesIgnite/crypto11"
	"github.com/btcsuite/btcd/btcec"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
	"math/big"
)

var secp256k1halfN = new(big.Int).Rsh(btcec.S256().N, 1)

type RemoteSignerPV struct {
	s crypto11.Signer
	logger log.Logger

	pubKeyCache tmcrypto.PubKey
}

func NewRemoteSignerPV(s crypto11.Signer, l log.Logger) *RemoteSignerPV {
	return &RemoteSignerPV{
		s,
		l,
		nil,
	}
}

/* TODO: WIP */
func (pv *RemoteSignerPV) GetPubKey() tmcrypto.PubKey {
	if pv.pubKeyCache != nil {
		return pv.pubKeyCache
	}
	signer := pv.s
	pk0 := signer.Public()
	if pk, ok := pk0.(*ecdsa.PublicKey); ok {
		p, err := PublicKeyToPubKeySecp256k1(pk)
		if err != nil {
			panic(err)
		}
		pv.logger.Debug("GetPubKey", "address", p.Address(), "pubkey", hex.EncodeToString(p[:]))
		pv.pubKeyCache = p
		return p
	} else {
		panic("invalid signer")
	}
}

func (pv *RemoteSignerPV) SignVote(chainID string, vote *types.Vote) error {
	pv.logger.Debug("SignVote", "chainID", chainID)
	if sigBytes, err := pv.signMsg(vote.SignBytes(chainID)); err != nil {
		return err
	} else {
		pv.logger.Debug("SignVote", "chainID", chainID, "sig", hex.EncodeToString(sigBytes), "sig_len", len(sigBytes))
		vote.Signature = sigBytes
		v := pv.GetPubKey().VerifyBytes(vote.SignBytes(chainID), sigBytes)
		pv.logger.Debug("VerifyBytes", "v", v)
		return nil
	}
}

func (pv *RemoteSignerPV) SignProposal(chainID string, proposal *types.Proposal) error {
	pv.logger.Debug("SignProposal", "chainID", chainID)
	if sigBytes, err := pv.signMsg(proposal.SignBytes(chainID)); err != nil {
		return err
	} else {
		pv.logger.Debug("SignProposal", "chainID", chainID, "sig", hex.EncodeToString(sigBytes), "sig_len", len(sigBytes))
		proposal.Signature = sigBytes
		v := pv.GetPubKey().VerifyBytes(proposal.SignBytes(chainID), sigBytes)
		pv.logger.Debug("VerifyBytes", "v", v)
		return nil
	}
}

func (pv *RemoteSignerPV) signMsg(msgBytes []byte) ([]byte, error) {
	hash := tmcrypto.Sha256(msgBytes)
	if derSig, err := pv.s.Sign(rand.Reader, hash[:], nil); err != nil {
		return nil, err
	} else {
		signature, err := btcec.ParseDERSignature(derSig, gincocrypto.Secp256k1()/*btcec.S256()*/)
		if err != nil {
			return nil, err
		}

		// TODO: for debug
		ecdsaPubkey := pv.s.Public().(*ecdsa.PublicKey)
		if !ecdsa.Verify(ecdsaPubkey, hash[:], signature.R, signature.S) {
			return nil, errors.New("failed to verify")
		}

		pv.logger.Debug("signature", "r", signature.R, "s", signature.S)
		rbytes, sbytes := signature.R.Bytes(), signature.S.Bytes()
		sigBytes := make([]byte, 64)
		copy(sigBytes[32-len(rbytes):32], rbytes)
		copy(sigBytes[64-len(sbytes):64], sbytes)

		// Reject malleable signatures.
		// Check tendermint@v0.32.3/crypto/secp256k1/secp256k1_nocgo.go for detail
		if signature.S.Cmp(secp256k1halfN) > 0 {
			return pv.signMsg(msgBytes)
		}

		// TODO: for debug
		if !pv.GetPubKey().VerifyBytes(msgBytes, sigBytes) {
			return nil, errors.New("failed to verify (2)")
		}

		return sigBytes, nil
	}
}

func makeRecoverableSignature(msg, sig []byte, expectedPubkey tmcrypto.PubKey) ([]byte, error) {
	for v := 0; v < 2; v++ {
		sig[64] = byte(v)
		if expectedPubkey.VerifyBytes(msg, sig) {
			return sig, nil
		}
	}
	return nil, errors.New("recovered public key mismatch")
}

func PublicKeyToPubKeySecp256k1(pubKey0 *ecdsa.PublicKey) (secp256k1.PubKeySecp256k1, error) {
	pubKey := btcec.PublicKey(*pubKey0)
	var tmPubkeyBytes secp256k1.PubKeySecp256k1
	copy(tmPubkeyBytes[:], pubKey.SerializeCompressed())
	return tmPubkeyBytes, nil
}
