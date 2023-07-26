package main

import (
	"dusnet/handler"
	"dusnet/logger"
	"dusnet/server"
)

func main() {
	var errCount int
	s1, err1 := startServer("mint_server1", "tcp", "0.0.0.0", 9000)
	if err1 != nil {
		logger.Error("tcp connect error,error:%+v", err1)
		errCount++
	}
	defer s1.Stop()

	s2, err2 := startServer("mint_server2", "tcp", "0.0.0.0", 9001)
	if err2 != nil {
		logger.Error("tcp connect error,error:%+v", err2)
		errCount++
	}
	defer s2.Stop()

	s3, err3 := startServer("mint_server3", "tcp", "0.0.0.0", 9002)
	if err3 != nil {
		logger.Error("tcp connect error,error:%+v", err3)
		errCount++
	}
	defer s3.Stop()
	if errCount == 3 {
		panic("all server shutdown already!,program quit")
	}
	select {}
}

func startServer(name, network, host string, port int) (server.IServer, error) {
	ping1000Handler := handler.Ping1000Handler{}
	// ping1000Handler.SetCodec(zcodec.Default())
	handler.RegisterChildHandler(1000, &ping1000Handler)
	s := server.Default(name, network, host, port)
	err := s.Start()
	if err != nil {
		logger.Error("server[%+v] start error", s)
	}
	return s, err
}
