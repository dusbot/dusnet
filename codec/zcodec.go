package zcodec

import (
	"bytes"
	"dusnet/logger"
	"dusnet/packet"
	"encoding/binary"
	"errors"
	"strconv"
)

const (
	TYPE_PING     = iota + 1234 // PING报文
	TYPE_SYNC                   // 同步报文
	TYPE_BUSINESS               // 业务报文 业务报文永远在队尾
)

type Icodec interface {
	Encode(packet.IPacket) ([]byte, error)
	Decode([]byte) (packet.IPacket, error)
}

func Default() Icodec {
	return &codec{}
}

type codec struct {
}

func (c *codec) Encode(p packet.IPacket) ([]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, 0))
	id := p.GetID()
	err := binary.Write(buffer, binary.BigEndian, &id)
	if err != nil {
		logger.Error("binary.Write head/id to bytes error,error:%+v", err)
		return buffer.Bytes(), err
	}
	type0 := p.GetType()
	err = binary.Write(buffer, binary.BigEndian, &type0)
	if err != nil {
		logger.Error("binary.Write head/type to bytes error,error:%+v", err)
		return buffer.Bytes(), err
	}
	bodyLen := p.GetBodyLen()
	err = binary.Write(buffer, binary.BigEndian, &bodyLen)
	if err != nil {
		logger.Error("binary.Write head/length to bytes error,error:%+v", err)
		return buffer.Bytes(), err
	}

	err = binary.Write(buffer, binary.BigEndian, p.GetData())
	if err != nil {
		logger.Error("binary.Write body/data to bytes error,error:%+v", err)
		return buffer.Bytes(), err
	}
	return buffer.Bytes(), nil
}

func (c *codec) Decode(raw []byte) (packet.IPacket, error) {
	pkt := &packet.Packet{}
	// decode head/id
	err := binary.Read(bytes.NewBuffer(raw[:4]), binary.BigEndian, &pkt.ID)
	if err != nil {
		logger.Error("binary.Read head/id to pkt error,error:%+v", err)
		return pkt, err
	}
	if pkt.ID == 0 {
		return pkt, errors.New("id zero value")
	}

	err = binary.Read(bytes.NewBuffer(raw[4:6]), binary.BigEndian, &pkt.Type)
	if err != nil {
		logger.Error("binary.Read head/type to pkt error,error:%+v", err)
		return pkt, err
	}
	if pkt.Type < TYPE_PING || pkt.Type > TYPE_BUSINESS {
		return pkt, errors.New("pkt type " + strconv.Itoa(int(pkt.Type)) + " not defined")
	}

	err = binary.Read(bytes.NewBuffer(raw[6:10]), binary.BigEndian, &pkt.Length)
	if err != nil {
		logger.Error("binary.Write head/length to pkt error,error:%+v", err)
		return pkt, err
	}

	if err != nil {
		logger.Error("conn.Read body/data error,error:%+v", err)
		return pkt, err
	}
	pkt.Data = raw[10 : 10+pkt.Length]
	return pkt, nil
}
