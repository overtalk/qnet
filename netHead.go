package qnet

import (
	"encoding/binary"
	"errors"
)

type HeadLength uint32
type HeadDeserializeFunc func(data []byte) (NetHead, error)
type HeadSerializeFunc func(head NetHead) []byte

const (
	CSHeadLength HeadLength = 6  // cs head
	SSHeadLength HeadLength = 22 // ss head
)

// NetHead define a net message head
type NetHead interface {
	GetMsgID() uint16
	GetMsgLength() uint32
	GetActorID() uint64
	GetSrcBusID() uint32
	GetDstBusID() uint32

	SetMsgID(value uint16)
	SetMsgLength(value uint32)
	SetActorID(value uint64)
	SetSrcBusID(value uint32)
	SetDstBusID(value uint32)
}

type BasicNetHead struct{}

func (head *BasicNetHead) GetMsgID() uint16     { return 0 }
func (head *BasicNetHead) GetMsgLength() uint32 { return 0 }
func (head *BasicNetHead) GetActorID() uint64   { return 0 }
func (head *BasicNetHead) GetSrcBusID() uint32  { return 0 }
func (head *BasicNetHead) GetDstBusID() uint32  { return 0 }

func (head *BasicNetHead) SetMsgID(value uint16)     {}
func (head *BasicNetHead) SetMsgLength(value uint32) {}
func (head *BasicNetHead) SetActorID(value uint64)   {}
func (head *BasicNetHead) SetSrcBusID(value uint32)  {}
func (head *BasicNetHead) SetDstBusID(value uint32)  {}

/*
| msg id | msg len |
|    2   |    4    | = 6
*/
type CSMsgHead struct {
	id     uint16 // msg id
	length uint32 // msg length (without header length)
}

func CSMsgHeadDeserializer(data []byte) (NetHead, error) {
	if len(data) != int(CSHeadLength) {
		return nil, errors.New("invalid cs header length")
	}

	return &CSMsgHead{
		id:     binary.BigEndian.Uint16(data[:2]),
		length: binary.BigEndian.Uint32(data[2:CSHeadLength]),
	}, nil
}

func CSMsgHeadSerializer(head NetHead) []byte {
	buf := make([]byte, head.GetMsgLength())
	binary.BigEndian.PutUint16(buf[:2], head.GetMsgID())
	binary.BigEndian.PutUint32(buf[2:], head.GetMsgLength())
	return buf
}

func (head *CSMsgHead) GetMsgID() uint16     { return head.id }
func (head *CSMsgHead) GetMsgLength() uint32 { return head.length }
func (head *CSMsgHead) GetActorID() uint64   { return 0 }
func (head *CSMsgHead) GetSrcBusID() uint32  { return 0 }
func (head *CSMsgHead) GetDstBusID() uint32  { return 0 }

func (head *CSMsgHead) SetMsgID(value uint16)     { head.id = value }
func (head *CSMsgHead) SetMsgLength(value uint32) { head.length = value }
func (head *CSMsgHead) SetActorID(value uint64)   {}
func (head *CSMsgHead) SetSrcBusID(value uint32)  {}
func (head *CSMsgHead) SetDstBusID(value uint32)  {}

/*
| msg id | msg len | actor id | src bus | dst bus |
|    2   |    4    |     8    |    4    |    4    | = 22
*/
type SSMsgHead struct {
	id       uint16 // msg id
	length   uint32 // msg length (without header length)
	actorID  uint64
	srcBusID uint32
	dstBusID uint32
}

func SSMsgHeadDeserializer(data []byte) (NetHead, error) {
	if len(data) != int(SSHeadLength) {
		return nil, errors.New("invalid cs header length")
	}

	return &SSMsgHead{
		id:       binary.BigEndian.Uint16(data[:2]),
		length:   binary.BigEndian.Uint32(data[2:CSHeadLength]),
		actorID:  binary.BigEndian.Uint64(data[CSHeadLength : CSHeadLength+8]),
		srcBusID: binary.BigEndian.Uint32(data[CSHeadLength+8 : CSHeadLength+12]),
		dstBusID: binary.BigEndian.Uint32(data[CSHeadLength+12:]),
	}, nil
}

func SSMsgHeadSerializer(head NetHead) []byte {
	buf := make([]byte, head.GetMsgLength())
	binary.BigEndian.PutUint16(buf[:2], head.GetMsgID())
	binary.BigEndian.PutUint32(buf[2:CSHeadLength], head.GetMsgLength())
	binary.BigEndian.PutUint64(buf[CSHeadLength:CSHeadLength+8], head.GetActorID())
	binary.BigEndian.PutUint32(buf[CSHeadLength+8:CSHeadLength+12], head.GetSrcBusID())
	binary.BigEndian.PutUint32(buf[CSHeadLength+12:], head.GetDstBusID())
	return buf
}

func (head *SSMsgHead) GetMsgID() uint16     { return head.id }
func (head *SSMsgHead) GetMsgLength() uint32 { return head.length }
func (head *SSMsgHead) GetActorID() uint64   { return head.actorID }
func (head *SSMsgHead) GetSrcBusID() uint32  { return head.srcBusID }
func (head *SSMsgHead) GetDstBusID() uint32  { return head.dstBusID }

func (head *SSMsgHead) SetMsgID(value uint16)     { head.id = value }
func (head *SSMsgHead) SetMsgLength(value uint32) { head.length = value }
func (head *SSMsgHead) SetActorID(value uint64)   { head.actorID = value }
func (head *SSMsgHead) SetSrcBusID(value uint32)  { head.srcBusID = value }
func (head *SSMsgHead) SetDstBusID(value uint32)  { head.dstBusID = value }
