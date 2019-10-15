package conn

import (
	amino "github.com/tendermint/go-amino"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/p2p/conn"
)

var cdc *amino.Codec = amino.NewCodec()

func init() {
	cryptoAmino.RegisterAmino(cdc)
	RegisterPacket(cdc)
}

func RegisterPacket(cdc *amino.Codec) {
	cdc.RegisterInterface((*conn.Packet)(nil), nil)
	cdc.RegisterConcrete(conn.PacketPing{}, "tendermint/p2p/PacketPing", nil)
	cdc.RegisterConcrete(conn.PacketPong{}, "tendermint/p2p/PacketPong", nil)
	cdc.RegisterConcrete(conn.PacketMsg{}, "tendermint/p2p/PacketMsg", nil)
}
