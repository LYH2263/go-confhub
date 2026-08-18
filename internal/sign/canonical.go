package sign

import (
	"bytes"
	"encoding/hex"
	"strconv"
)

const (
	AlgoHMAC    = "hmac-sha256"
	AlgoEd25519 = "ed25519"
)

// Canonical 稳定的待签名字节：版本行 + ns + key + 可选 rev + payload。
// 不含签名本身，避免循环。
func Canonical(ns, key string, rev int64, payload []byte) []byte {
	var b bytes.Buffer
	b.Grow(32 + len(ns) + len(key) + len(payload))
	b.WriteString("confhub-sign-v1\n")
	b.WriteString("ns=")
	b.WriteString(ns)
	b.WriteByte('\n')
	b.WriteString("key=")
	b.WriteString(key)
	b.WriteByte('\n')
	if rev > 0 {
		b.WriteString("rev=")
		b.WriteString(strconv.FormatInt(rev, 10))
		b.WriteByte('\n')
	}
	b.WriteString("payload-len=")
	b.WriteString(strconv.Itoa(len(payload)))
	b.WriteByte('\n')
	b.WriteString("payload-sha256=")
	b.WriteString(hex.EncodeToString(sha256Sum(payload)))
	b.WriteByte('\n')
	return b.Bytes()
}

func CanonicalNoRev(ns, key string, payload []byte) []byte {
	return Canonical(ns, key, 0, payload)
}

func NormalizeAlgo(algo string) string {
	switch algo {
	case "", AlgoHMAC, "hmac", "sha256":
		if algo == "" {
			return ""
		}
		return AlgoHMAC
	case AlgoEd25519, "ed":
		return AlgoEd25519
	default:
		return algo
	}
}
