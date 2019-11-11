package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"github.com/ThalesIgnite/crypto11"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/types"
)

type RemoteSignerPV struct {
	s crypto11.Signer
	logger log.Logger
}

func NewRemoteSignerPV(s crypto11.Signer, l log.Logger) *RemoteSignerPV {
	return &RemoteSignerPV{
		s,
		l,
	}
}

/* TODO: WIP */
func (pv *RemoteSignerPV) GetPubKey() tmcrypto.PubKey {
	signer := pv.s
	pk0 := signer.Public()
	if pk, ok := pk0.(*ecdsa.PublicKey); ok {
		var p secp256k1.PubKeySecp256k1
		copy(p[:], pk.X.Bytes())
		pv.logger.Debug("GetPubKey", "pubkey", hex.EncodeToString(p[:]))
		return p
	} else {
		panic("invalid signer")
	}
}

func (pv *RemoteSignerPV) SignVote(chainID string, vote *types.Vote) error {
	if sig, err := pv.s.Sign(rand.Reader, vote.SignBytes(chainID), nil); err != nil {
		return err
	} else {
		pv.logger.Debug("SignVote", "chainID", chainID)
		vote.Signature = sig
		return nil
	}
}

func (pv *RemoteSignerPV) SignProposal(chainID string, proposal *types.Proposal) error {
	if sig, err := pv.s.Sign(rand.Reader, proposal.SignBytes(chainID), nil); err != nil {
		return err
	} else {
		pv.logger.Debug("SignProposal", "chainID", chainID)
		proposal.Signature = sig
		return nil
	}
}
