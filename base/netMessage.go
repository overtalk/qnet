package base

type NetMsg struct {
	head    NetHead
	msgData []byte
}

func NewNetMsg(head NetHead, data []byte) *NetMsg {
	netMsg := &NetMsg{
		head:    head,
		msgData: make([]byte, len(data)),
	}

	// copy
	copy(netMsg.msgData, data)
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
