package snapshot

import (
	"encoding/binary"
	"hash/crc32"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/meta"
)

func putUvarint(b *[]byte, v uint64) {
	var tmp [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(tmp[:], v)
	*b = append(*b, tmp[:n]...)
}

func putBytes(b *[]byte, p []byte) {
	putUvarint(b, uint64(len(p)))
	*b = append(*b, p...)
}

func putStr(b *[]byte, s string) { putBytes(b, []byte(s)) }

func putI64(b *[]byte, v int64) { putUvarint(b, uint64(v)) }

func putTime(b *[]byte, t interface{ UnixNano() int64 }) {
	putI64(b, t.UnixNano())
}

func putGray(b *[]byte, g *meta.GrayRule) {
	if g == nil {
		putUvarint(b, 0)
		return
	}
	putUvarint(b, 1)
	putI64(b, int64(g.Percent))
	putUvarint(b, uint64(len(g.AllowIDs)))
	for _, id := range g.AllowIDs {
		putStr(b, id)
	}
	putUvarint(b, uint64(len(g.MatchTags)))
	for k, v := range g.MatchTags {
		putStr(b, k)
		putStr(b, v)
	}
}

func Encode(blob *Blob) ([]byte, error) {
	if blob == nil {
		return nil, cherr.ErrSnapshot
	}
	blob.Tally()
	var body []byte
	putStr(&body, blob.Header.Magic)
	putI64(&body, int64(blob.Header.Format))
	putTime(&body, blob.Header.CreatedAt)

	putUvarint(&body, uint64(len(blob.Namespaces)))
	for _, n := range blob.Namespaces {
		putStr(&body, n.ID)
		putStr(&body, n.Owner)
		putTime(&body, n.CreatedAt)
		putI64(&body, int64(n.MaxKeys))
		putI64(&body, n.MaxBytes)
	}

	putUvarint(&body, uint64(len(blob.Entries)))
	for _, e := range blob.Entries {
		if e == nil {
			return nil, cherr.Wrap(cherr.ErrSnapshot, "nil entry")
		}
		putStr(&body, e.NS)
		putStr(&body, e.Key)
		putI64(&body, e.Head)
		putI64(&body, e.Stable)
		putUvarint(&body, uint64(len(e.Versions)))
		for _, v := range e.Versions {
			putI64(&body, v.Rev)
			putBytes(&body, v.Payload)
			putBytes(&body, v.Sig)
			putStr(&body, v.Algo)
			putStr(&body, v.Author)
			putTime(&body, v.Ts)
			putGray(&body, v.Gray)
		}
	}

	putUvarint(&body, uint64(len(blob.Audit)))
	for _, a := range blob.Audit {
		putI64(&body, a.Seq)
		putTime(&body, a.Ts)
		putStr(&body, a.NS)
		putStr(&body, a.Key)
		putI64(&body, a.Rev)
		putStr(&body, a.Kind)
		putStr(&body, a.Actor)
		putStr(&body, a.Detail)
		putI64(&body, int64(a.Bytes))
		if a.OK {
			putUvarint(&body, 1)
		} else {
			putUvarint(&body, 0)
		}
		putStr(&body, a.Err)
	}

	putUvarint(&body, uint64(len(blob.Grants)))
	for ns, subs := range blob.Grants {
		putStr(&body, ns)
		putUvarint(&body, uint64(len(subs)))
		for sub, role := range subs {
			putStr(&body, sub)
			putStr(&body, role)
		}
	}

	putUvarint(&body, uint64(len(blob.Usage)))
	for ns, u := range blob.Usage {
		putStr(&body, ns)
		putI64(&body, int64(u.Keys))
		putI64(&body, u.Bytes)
	}

	sum := crc32.ChecksumIEEE(body)
	out := make([]byte, 0, 4+len(body)+4)
	out = append(out, Magic...)
	out = append(out, body...)
	var crc [4]byte
	binary.BigEndian.PutUint32(crc[:], sum)
	out = append(out, crc[:]...)
	return out, nil
}
