package base

import (
	"encoding/binary"
	"errors"
)

type HeadLength uint32

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

/*
| msg id | msg len |
|    2   |    4    | = 6
*/
type CSMsgHead struct {
	id     uint16 // msg id
	length uint32 // msg length (without header length)
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

////////////////////////////////////////////////////////
// net message
////////////////////////////////////////////////////////

type NetMsg struct {
	head    NetHead
	msgData []byte
}

func NewSSNetMsgFromData(data []byte) *NetMsg {
	netMsg := &NetMsg{
		head: &SSMsgHead{
			length: uint32(len(data)),
		},
		msgData: make([]byte, len(data)),
	}

	// copy
	copy(netMsg.msgData, data)
	return netMsg
}

func NewSSNetMsgFromSSNetMsg(msg *NetMsg) *NetMsg {
	netMsg := NewSSNetMsgFromData(msg.msgData)
	netMsg.head = msg.head
	return netMsg
}

// get
func (netMsg *NetMsg) GetHead() NetHead { return netMsg.head }
func (netMsg *NetMsg) GetMsg() []byte   { return netMsg.msgData }

// use the func of head, just for easy to use
func (netMsg *NetMsg) GetMsgID() uint16          { return netMsg.head.GetMsgID() }
func (netMsg *NetMsg) GetMsgLength() uint32      { return netMsg.head.GetMsgLength() }
func (netMsg *NetMsg) GetActorID() uint64        { return netMsg.head.GetActorID() }
func (netMsg *NetMsg) GetSrcBusID() uint32       { return netMsg.head.GetSrcBusID() }
func (netMsg *NetMsg) GetDstBusID() uint32       { return netMsg.head.GetDstBusID() }
func (netMsg *NetMsg) SetMsgID(value uint16)     { netMsg.head.SetMsgID(value) }
func (netMsg *NetMsg) SetMsgLength(value uint32) { netMsg.head.SetMsgLength(value) }
func (netMsg *NetMsg) SetActorID(value uint64)   { netMsg.head.SetActorID(value) }
func (netMsg *NetMsg) SetSrcBusID(value uint32)  { netMsg.head.SetSrcBusID(value) }
func (netMsg *NetMsg) SetDstBusID(value uint32)  { netMsg.head.SetDstBusID(value) }

////////////////////////////////////////////////////////
// deserialize net message
////////////////////////////////////////////////////////
func DeserializeMsgHead(l HeadLength, data []byte) (NetHead, error) {
	if len(data) != int(l) {
		return nil, errors.New("invalid header length")
	}

	switch l {
	case CSHeadLength:
		return &CSMsgHead{
			id:     binary.BigEndian.Uint16(data[:2]),
			length: binary.BigEndian.Uint32(data[2:CSHeadLength]),
		}, nil
	case SSHeadLength:
		return &SSMsgHead{
			id:       binary.BigEndian.Uint16(data[:2]),
			length:   binary.BigEndian.Uint32(data[2:CSHeadLength]),
			actorID:  binary.BigEndian.Uint64(data[CSHeadLength : CSHeadLength+8]),
			srcBusID: binary.BigEndian.Uint32(data[CSHeadLength+8 : CSHeadLength+12]),
			dstBusID: binary.BigEndian.Uint32(data[CSHeadLength+12:]),
		}, nil
	default:
		return nil, errors.New("unknown HeadLength")
	}
}
