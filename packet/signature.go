package packet

import (
	"crypto/hmac"
	"crypto/sha1"
)

var defaultSignSecret []byte

// SetSignSecret set the signature secret
func SetSignSecret(sec []byte) {
	defaultSignSecret = append([]byte{}, sec...)
}

// ISignature a signature to calculate the sum
type ISignature interface {
	Sum(token, data []byte) ([]byte, error)
}

type hmacSha1Signature struct{}

// HMACSha1Signature a hmac-sha1 signature
var HMACSha1Signature ISignature = hmacSha1Signature{}

// Sum calculate the signature of the dataload
func (hmacSha1Signature) Sum(token, data []byte) ([]byte, error) {
	hmac := hmac.New(sha1.New, append(defaultSignSecret, token...))
	_, err := hmac.Write(data)
	if err == nil {
		return hmac.Sum(nil), nil
	}
	return nil, err
}
