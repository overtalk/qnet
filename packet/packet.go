package packet

import (
	"encoding/binary"
	"errors"
)

// IPacket a packet abstract
type IPacket interface {
	GetDataSize() uint16
	SetDataSize(size uint16)
	GetConnID() uint32
	SetConnID(id uint32)
	GetProtoID() uint16
	SetProtoID(id uint16)
	GetProtoVer() uint8
	SetProtoVer(ver uint8)
	GetDataFlag() uint8
	SetDataFlag(flag uint8)
	GetDataSign() []byte
	SetDataSign(sign []byte)
	GetDataLoad() []byte
	SetDataLoad(data []byte)
}

// error definitions
var ErrInvalidSize = errors.New("invalid packet size")

const (
	// packet size
	OptSizeCmd    = 6
	OptSizeData   = 8
	MaxPacketSize = 32 * 1024

	// data flags
	FlagZLIB     = 0x01
	FlagXOR      = 0x02
	FlagHMACSha1 = 0x04

	// cmd id
	CmdPing     = 0x0000
	CmdRegister = 0x0001
)

// Packet a agent protocol
type Packet []byte

// New create a Packet
func New(datasize uint16) Packet {
	pack := Packet(make([]byte, 2+datasize))
	pack.SetDataSize(datasize)
	return pack
}

// Check check whether it's a valid packet
func Check(b []byte) (uint16, error) {
	dataLen := len(b)
	if dataLen < OptSizeCmd || dataLen > MaxPacketSize {
		return 0, ErrInvalidSize
	}
	dataSize := binary.BigEndian.Uint16(b[:2])
	if dataSize != uint16(dataLen)-2 {
		return 0, ErrInvalidSize
	}
	return dataSize, nil
}

// NewFromBytes create a Packet from a raw bytes
func NewFromBytes(b []byte) (Packet, error) {
	dataSize, err := Check(b)
	return Packet(b[:dataSize]), err
}

// NewFromData create a Packet from a data
func NewFromData(data, sign []byte, compressor ICompresser) Packet {
	var compressed bool
	data, compressed = compressor.Compress(data)
	if signSize := len(sign); signSize > 0 {
		packet := New(uint16(OptSizeData + 1 + signSize + len(data)))
		packet.SetZlibCompressed(compressed)
		packet.SetDataFlag(FlagHMACSha1)
		packet.SetDataSign(sign)
		packet.SetDataLoad(data)
		compressor.Close()
		return packet
	}
	packet := New(uint16(OptSizeData + len(data)))
	packet.SetZlibCompressed(compressed)
	packet.SetDataLoad(data)
	compressor.Close()
	return packet
}

// NewPing create a PingPacket
// which is DATASIZE + CONNID + PROTOID
func NewPing() Packet {
	packet := New(OptSizeCmd)
	packet.SetConnID(0)
	packet.SetProtoID(0)
	return packet
}

// PingPacket the default ping packet
var PingPacket = NewPing()

// NewRegister create a RegisterPacket
// which is DATASIZE + CONNID + PROTOID
func NewRegister(sid uint32) Packet {
	packet := New(OptSizeCmd)
	packet.SetConnID(sid)
	packet.SetProtoID(1)
	return packet
}

// MakeProtoID make a proto id by the mid and aid
func MakeProtoID(mid, aid uint8) uint16 {
	return uint16(mid)<<8 + uint16(aid)
}

// SplitProtoID split the proto id to the mid and aid
func SplitProtoID(protoID uint16) (uint8, uint8) {
	return uint8((protoID >> 8) & 0x00FF), uint8(protoID & 0x00FF)
}

// IsValid check whether the packet is valid
func (packet Packet) IsValid() bool {
	return len(packet) >= (2 + OptSizeCmd)
}

// IsCmdSize check whether is a cmd packet's size
func (packet Packet) IsCmdSize() bool { return packet.GetDataSize() == OptSizeCmd }
func (packet Packet) GetCmd() uint16  { return binary.BigEndian.Uint16(packet[6:8]) }

// IsCmdProto check whether it's a cmd proto
func (packet Packet) IsCmdProto() bool {
	// cmd proto[6-7]: 0x0000 ~ 0x00FF
	return packet[6] == 0
}

func (packet Packet) Encrypt(crypto ICrypto)      { crypto.Encrypt(packet) }
func (packet Packet) Decrypt(crypto ICrypto)      { crypto.Decrypt(packet) }
func (packet Packet) GetDataSize() uint16         { return binary.BigEndian.Uint16(packet[:2]) }
func (packet Packet) SetDataSize(datasize uint16) { binary.BigEndian.PutUint16(packet[:2], datasize) }
func (packet Packet) GetConnID() uint32           { return binary.BigEndian.Uint32(packet[2:6]) }
func (packet Packet) SetConnID(id uint32)         { binary.BigEndian.PutUint32(packet[2:6], id) }
func (packet Packet) GetProtoID() uint16          { return binary.BigEndian.Uint16(packet[6:8]) }
func (packet Packet) SetProtoID(protoID uint16)   { binary.BigEndian.PutUint16(packet[6:8], protoID) }
func (packet Packet) GetProtoMID() uint8          { return packet[6] }
func (packet Packet) SetProtoMID(mid uint8)       { packet[6] = mid }
func (packet Packet) GetProtoAID() uint8          { return packet[7] }
func (packet Packet) SetProtoAID(aid uint8)       { packet[7] = aid }
func (packet Packet) GetProtoVer() uint8          { return packet[8] }
func (packet Packet) SetProtoVer(version uint8)   { packet[8] = version }
func (packet Packet) GetDataFlag() uint8          { return packet[9] }
func (packet Packet) SetDataFlag(flag uint8)      { packet[9] |= flag }

// HasDataFlag check whether it has a data flag
func (packet Packet) HasDataFlag(flag uint8) bool {
	return packet[9]&flag == flag
}

// ClearDataFlag clear the data flag
func (packet Packet) ClearDataFlag(flag uint8) {
	// FIXME: cannot clear the signature bits, a mark will be ok
	packet[9] &^= flag
}

// ResetDataFlag reset the data flag
func (packet Packet) ResetDataFlag(flag uint8) {
	packet[9] = flag
}

// HasDataSign check whether it's a signature
func (packet Packet) HasDataSign() bool {
	return packet[9]&0x0C != 0
}

// SetZlibCompressed set the data flag: ZLIB
func (packet Packet) SetZlibCompressed(compressed bool) {
	if compressed {
		packet.SetDataFlag(FlagZLIB)
	}
}

// IsZlibCompressed check whether it's zlib-compressed
func (packet Packet) IsZlibCompressed() bool {
	return packet.HasDataFlag(FlagZLIB)
}

// GetDataSign get the signature of dataload
func (packet Packet) GetDataSign() []byte {
	size := int(packet[2+OptSizeData])
	return packet[3+OptSizeData : 3+OptSizeData+size]
}

// SetDataSign set the signature of dataload
func (packet Packet) SetDataSign(sign []byte) {
	packet[2+OptSizeData] = byte(len(sign))
	copy(packet[3+OptSizeData:], sign)
}

func (packet Packet) getDataLoadIndex() int {
	var index = 2 + OptSizeData
	if packet.HasDataSign() {
		index = 3 + OptSizeData + int(packet[2+OptSizeData])
	}
	return index
}

// GetDataLoad get the packet's dataload
func (packet Packet) GetDataLoad() []byte {
	index := packet.getDataLoadIndex()
	return packet[index:]
}

// SetDataLoad set the packet's dataload
func (packet Packet) SetDataLoad(data []byte) {
	index := packet.getDataLoadIndex()
	copy(packet[index:], data)
}
