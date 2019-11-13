package remotepv

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/ThalesIgnite/crypto11"
	"github.com/btcsuite/btcd/btcec"
	// gethcrypto "github.com/ethereum/go-ethereum/crypto"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

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
	if derSig, err := pv.s.Sign(rand.Reader, msgBytes, nil); err != nil {
		return nil, err
	} else {
		signature, err := btcec.ParseDERSignature(derSig, btcec.S256())
		if err != nil {
			return nil, err
		}

		// TODO: for debug
		ecdsaPubkey := pv.s.Public().(*ecdsa.PublicKey)
		if !ecdsa.Verify(ecdsaPubkey, msgBytes, signature.R, signature.S) {
			return nil, errors.New("failed to verify")
		}

		rbytes, sbytes := signature.R.Bytes(), signature.S.Bytes()
		sigBytes := make([]byte, 65)
		copy(sigBytes[32-len(rbytes):32], rbytes)
		copy(sigBytes[64-len(sbytes):64], sbytes)

		pubkey := pv.GetPubKey()
		return makeRecoverableSignature(msgBytes, sigBytes, pubkey)
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
