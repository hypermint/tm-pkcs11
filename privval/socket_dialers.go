package privval

import (
	"github.com/tendermint/tendermint/privval"
	"net"
	"time"

	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
	pkconn "github.com/datachainlab/tm-pkcs11/conn"
)

// DialTCPFn dials the given tcp addr, using the given timeoutReadWrite and
// privKey for the authenticated encryption handshake.
func DialTCPFn(addr string, timeoutReadWrite time.Duration, privKey crypto.PrivKey) privval.SocketDialer {
	return func() (net.Conn, error) {
		conn, err := cmn.Connect(addr)
		if err == nil {
			deadline := time.Now().Add(timeoutReadWrite)
			err = conn.SetDeadline(deadline)
		}
		if err == nil {
			conn, err = pkconn.MakeSecretConnection(conn, privKey)
		}
		return conn, err
	}
}

// DialUnixFn dials the given unix socket.
func DialUnixFn(addr string) privval.SocketDialer {
	return func() (net.Conn, error) {
		unixAddr := &net.UnixAddr{Name: addr, Net: "unix"}
		return net.DialUnix("unix", nil, unixAddr)
	}
}
