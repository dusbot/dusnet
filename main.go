package main

import (
	"dusnet/handler"
	"dusnet/logger"
	"dusnet/server"
)

func main() {
	s1 := startServer("mint_server1", "tcp", "0.0.0.0", 9000)
	defer s1.Stop()

	// s2 := startServer("mint_server2", "tcp", "0.0.0.0", 9001)
	// defer s2.Stop()

	// s3 := startServer("mint_server3", "tcp", "0.0.0.0", 9002)
	// defer s3.Stop()

	select {}
}

func startServer(name, network, host string, port int) server.IServer {
	ping1000Handler := handler.Ping1000Handler{}
	// ping1000Handler.SetCodec(zcodec.Default())
	handler.RegisterChildHandler(1000, &ping1000Handler)
	s := server.Default(name, network, host, port)
	err := s.Start()
	if err != nil {
		logger.Error("server[%+v] start error", s)
	}
	return s
}
