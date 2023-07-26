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

func main() {
	data := []byte(uuid.New().String())
	send(&packet.Packet{
		PacketHead: packet.PacketHead{
			ID:     1000,
			Type:   1234,
			Length: uint32(len(data)),
		},
		PacketBody: packet.PacketBody{Data: data},
	}, 9000)
	// for i := 0; i < 100; i++ {
	// 	go func() {
	// 		for {
	// 			dialTest()
	// 			time.Sleep(time.Second * 3)
	// 		}
	// 	}()
	// }
	// select {}
}

func send(pkt packet.IPacket, port ...int) {
	var saddr string
	if port == nil || port[0] == 0 {
		saddr = fmt.Sprintf("0.0.0.0:%d", rand.Intn(3)+9000)
	} else {
		saddr = fmt.Sprintf("0.0.0.0:%d", port[0])
	}
	dial, err := net.Dial("tcp", saddr)
	if err != nil {
		logger.Error("tcp connect to %s error,error:%+v", saddr, err)
		return
	}
	defer dial.Close()
	buf, err := zcodec.Default().Encode(pkt)
	if err != nil {
		panic(err)
	}
	dial.Write(buf)
}
