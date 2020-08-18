package packet

var defaultCryptoSecret []byte

// SetCryptoSecret set the xor secret
func SetCryptoSecret(sec []byte) {
	defaultCryptoSecret = append([]byte{}, sec...)
}

// ICrypto a crypto to encrypt/decrypt a packet
type ICrypto interface {
	Encrypt(Packet)
	Decrypt(Packet)
}

type xorCrypto struct{}

// XORCrypto a XOR crypto
var XORCrypto ICrypto = &xorCrypto{}

func (*xorCrypto) fixSecret(secret []byte) {
	var secIdx int
	if secret[0] == 0 {
		secret[0] = defaultCryptoSecret[0]
		secIdx = 1
	}
	if secret[1] == 0 {
		secret[1] = defaultCryptoSecret[secIdx]
	}
}

func (xor *xorCrypto) encryptOrDecryptOptvals(packet Packet) {
	dataSize := packet.GetDataSize()
	secret := []byte{
		byte((dataSize >> 8) & 0xFF),
		byte(dataSize & 0xFF),
	}
	xor.fixSecret(secret)
	// don't encrypt DATAFLAG
	for i := 2; i < 9; i++ {
		packet[i] ^= secret[i&0x01]
	}
}

func (xor *xorCrypto) encryptOrDecryptDataLoad(packet Packet) {
	secret := []byte{packet.GetProtoMID(), packet.GetProtoAID()}
	xor.fixSecret(secret)
	dataLen := len(packet)
	for i := 10; i < dataLen; i++ {
		packet[i] ^= secret[i&0x01]
	}
}

func (xor *xorCrypto) Encrypt(packet Packet) {
	xor.encryptOrDecryptDataLoad(packet)
	xor.encryptOrDecryptOptvals(packet)
	packet.SetDataFlag(FlagXOR)
}

func (xor *xorCrypto) Decrypt(packet Packet) {
	if packet.HasDataFlag(FlagXOR) {
		xor.encryptOrDecryptOptvals(packet)
		xor.encryptOrDecryptDataLoad(packet)
		packet.ClearDataFlag(FlagXOR)
	}
}
