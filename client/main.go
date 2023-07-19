package main

import (
	zcodec "dusnet/codec"
	"dusnet/packet"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/google/uuid"
)

func main() {
	for i := 0; i < 100; i++ {
		go func() {
			for {
				dialTest()
				time.Sleep(time.Second * 3)
			}
		}()
	}
	select {}
}

func dialTest() {
	dial, err := net.Dial("tcp", fmt.Sprintf("0.0.0.0:%d", rand.Intn(3)+9000))
	if err != nil {
		return
	}
	ids := []uint32{1000, 2000, 3000}
	defer dial.Close()
	data := []byte(uuid.New().String())
	buf, err := zcodec.Default().Encode(&packet.Packet{
		PacketHead: packet.PacketHead{
			ID:     ids[rand.Intn(3)],
			Type:   1234,
			Length: uint32(len(data)),
		},
		PacketBody: packet.PacketBody{Data: data},
	})
	if err != nil {
		panic(err)
	}
	dial.Write(buf)
}
