package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func sha256Sum(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}

func HMACSign(secret, message []byte) []byte {
	m := hmac.New(sha256.New, secret)
	_, _ = m.Write(message)
	return m.Sum(nil)
}

func HMACVerify(secret, message, sig []byte) bool {
	if len(secret) == 0 || len(sig) == 0 {
		return false
	}
	want := HMACSign(secret, message)
	return hmac.Equal(want, sig)
}

func HMACSignHex(secret, message []byte) string {
	return hex.EncodeToString(HMACSign(secret, message))
}

func DecodeSig(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}
	return hex.DecodeString(s)
}
