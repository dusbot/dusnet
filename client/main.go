package main

import (
	zcodec "dusnet/codec"
	"dusnet/logger"
	"dusnet/packet"
	"fmt"
	"math/rand"
	"net"

	"github.com/google/uuid"
)

var ids = []uint32{1000, 2000, 3000}

func main() {
	data := []byte(uuid.New().String())
	var pkts []packet.IPacket
	for i := 0; i < 3; i++ {
		pkt := packet.Packet{
			PacketHead: packet.PacketHead{
				ID:     ids[i],
				Type:   1234,
				Length: uint32(len(data)),
			},
			PacketBody: packet.PacketBody{Data: data},
		}
		pkts = append(pkts, &pkt)
	}
	send(9000, pkts...)
}

func send(port int, pkts ...packet.IPacket) {
	var saddr string
	if port == 0 {
		saddr = fmt.Sprintf("0.0.0.0:%d", rand.Intn(3)+9000)
	} else {
		saddr = fmt.Sprintf("0.0.0.0:%d", port)
	}
	dial, err := net.Dial("tcp", saddr)
	if err != nil {
		logger.Error("tcp connect to %s error,error:%+v", saddr, err)
		return
	}
	defer dial.Close()
	for _, pkt := range pkts {
		buf, err := zcodec.Default().Encode(pkt)
		if err != nil {
			panic(err)
		}
		dial.Write(buf)
	}
}
