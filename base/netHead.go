package base

import (
	"encoding/binary"
	"errors"
)

type HeadLength uint32
type HeadDeserializeFunc func(data []byte) (NetHead, error)

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

type BaseNetHead struct{}

func (head *BaseNetHead) GetMsgID() uint16     { return 0 }
func (head *BaseNetHead) GetMsgLength() uint32 { return 0 }
func (head *BaseNetHead) GetActorID() uint64   { return 0 }
func (head *BaseNetHead) GetSrcBusID() uint32  { return 0 }
func (head *BaseNetHead) GetDstBusID() uint32  { return 0 }

func (head *BaseNetHead) SetMsgID(value uint16)     {}
func (head *BaseNetHead) SetMsgLength(value uint32) {}
func (head *BaseNetHead) SetActorID(value uint64)   {}
func (head *BaseNetHead) SetSrcBusID(value uint32)  {}
func (head *BaseNetHead) SetDstBusID(value uint32)  {}

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
