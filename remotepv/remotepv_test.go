package remotepv

import (
	"github.com/tendermint/tendermint/libs/log"
	"os"
	"testing"
)

func TestSignMsg(t *testing.T) {
	logger := log.NewTMLogger(
		log.NewSyncWriter(os.Stdout),
	).With("module", "signer")
	NewRemoteSignerPV(logger)
}
