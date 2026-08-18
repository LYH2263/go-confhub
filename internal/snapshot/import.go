package snapshot

import (
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"time"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/meta"
	"github.com/LYH2263/go-confhub/internal/quota"
)

type reader struct {
	b []byte
	i int
}

func (r *reader) uvarint() (uint64, error) {
	v, n := binary.Uvarint(r.b[r.i:])
	if n <= 0 {
		return 0, cherr.ErrSnapshot
	}
	r.i += n
	return v, nil
}

func (r *reader) i64() (int64, error) {
	v, err := r.uvarint()
	return int64(v), err
}

func (r *reader) bytes() ([]byte, error) {
	n, err := r.uvarint()
	if err != nil {
		return nil, err
	}
	if r.i+int(n) > len(r.b) {
		return nil, cherr.ErrSnapshot
	}
	p := r.b[r.i : r.i+int(n)]
	r.i += int(n)
	out := make([]byte, len(p))
	copy(out, p)
	return out, nil
}

func (r *reader) str() (string, error) {
	b, err := r.bytes()
	return string(b), err
}

func (r *reader) gray() (*meta.GrayRule, error) {
	flag, err := r.uvarint()
	if err != nil {
		return nil, err
	}
	if flag == 0 {
		return nil, nil
	}
	g := &meta.GrayRule{}
	p, err := r.i64()
	if err != nil {
		return nil, err
	}
	g.Percent = int(p)
	n, err := r.uvarint()
	if err != nil {
		return nil, err
	}
	g.AllowIDs = make([]string, 0, n)
	for i := uint64(0); i < n; i++ {
		s, err := r.str()
		if err != nil {
			return nil, err
		}
		g.AllowIDs = append(g.AllowIDs, s)
	}
	tn, err := r.uvarint()
	if err != nil {
		return nil, err
	}
	if tn > 0 {
		g.MatchTags = make(map[string]string, tn)
	}
	for i := uint64(0); i < tn; i++ {
		k, err := r.str()
		if err != nil {
			return nil, err
		}
		v, err := r.str()
		if err != nil {
			return nil, err
		}
		g.MatchTags[k] = v
	}
	return g, nil
}

func Decode(raw []byte) (*Blob, error) {
	if len(raw) < 8 {
		return nil, cherr.ErrSnapshot
	}
	if string(raw[:4]) != Magic {
		// 允许 JSON 快照
		var b Blob
		if err := json.Unmarshal(raw, &b); err != nil {
			return nil, cherr.Wrap(cherr.ErrSnapshot, "magic")
		}
		if b.Header.Magic == "" {
			b.Header.Magic = Magic
		}
		return &b, nil
	}
	body := raw[4 : len(raw)-4]
	want := binary.BigEndian.Uint32(raw[len(raw)-4:])
	got := crc32.ChecksumIEEE(body)
	if want != got {
		return nil, cherr.ErrChecksum
	}
	r := reader{b: body}
	magic, err := r.str()
	if err != nil {
		return nil, err
	}
	format, err := r.i64()
	if err != nil {
		return nil, err
	}
	created, err := r.i64()
	if err != nil {
		return nil, err
	}
	blob := NewBlob(time.Unix(0, created).UTC())
	blob.Header.Magic = magic
	blob.Header.Format = int(format)

	nn, err := r.uvarint()
	if err != nil {
		return nil, err
	}
	for i := uint64(0); i < nn; i++ {
		id, err := r.str()
		if err != nil {
			return nil, err
		}
		owner, err := r.str()
		if err != nil {
			return nil, err
		}
		ts, err := r.i64()
		if err != nil {
			return nil, err
		}
		mk, err := r.i64()
		if err != nil {
			return nil, err
		}
		mb, err := r.i64()
		if err != nil {
			return nil, err
		}
		blob.Namespaces = append(blob.Namespaces, meta.Namespace{
			ID: id, Owner: owner, CreatedAt: time.Unix(0, ts).UTC(),
			MaxKeys: int(mk), MaxBytes: mb,
		})
	}

	en, err := r.uvarint()
	if err != nil {
		return nil, err
	}
	for i := uint64(0); i < en; i++ {
		ns, err := r.str()
		if err != nil {
			return nil, err
		}
		key, err := r.str()
		if err != nil {
			return nil, err
		}
		head, err := r.i64()
		if err != nil {
			return nil, err
		}
		stable, err := r.i64()
		if err != nil {
			return nil, err
		}
		vn, err := r.uvarint()
		if err != nil {
			return nil, err
		}
		e := &meta.Entry{NS: ns, Key: key, Head: head, Stable: stable}
		for j := uint64(0); j < vn; j++ {
			rev, err := r.i64()
			if err != nil {
				return nil, err
			}
			payload, err := r.bytes()
			if err != nil {
				return nil, err
			}
			sig, err := r.bytes()
			if err != nil {
				return nil, err
			}
			algo, err := r.str()
			if err != nil {
				return nil, err
			}
			author, err := r.str()
			if err != nil {
				return nil, err
			}
			ts, err := r.i64()
			if err != nil {
				return nil, err
			}
			g, err := r.gray()
			if err != nil {
				return nil, err
			}
			e.Versions = append(e.Versions, meta.VersionMeta{
				Rev: rev, Payload: payload, Sig: sig, Algo: algo, Author: author,
				Ts: time.Unix(0, ts).UTC(), Gray: g,
			})
		}
		blob.Entries = append(blob.Entries, e)
	}

	an, err := r.uvarint()
	if err != nil {
		return nil, err
	}
	for i := uint64(0); i < an; i++ {
		seq, err := r.i64()
		if err != nil {
			return nil, err
		}
		ts, err := r.i64()
		if err != nil {
			return nil, err
		}
		ns, err := r.str()
		if err != nil {
			return nil, err
		}
		key, err := r.str()
		if err != nil {
			return nil, err
		}
		rev, err := r.i64()
		if err != nil {
			return nil, err
		}
		kind, err := r.str()
		if err != nil {
			return nil, err
		}
		actor, err := r.str()
		if err != nil {
			return nil, err
		}
		detail, err := r.str()
		if err != nil {
			return nil, err
		}
		bytesN, err := r.i64()
		if err != nil {
			return nil, err
		}
		okv, err := r.uvarint()
		if err != nil {
			return nil, err
		}
		errS, err := r.str()
		if err != nil {
			return nil, err
		}
		blob.Audit = append(blob.Audit, meta.AuditRecord{
			Seq: seq, Ts: time.Unix(0, ts).UTC(), NS: ns, Key: key, Rev: rev,
			Kind: kind, Actor: actor, Detail: detail, Bytes: int(bytesN), OK: okv == 1, Err: errS,
		})
	}

	gn, err := r.uvarint()
	if err != nil {
		return nil, err
	}
	blob.Grants = make(map[string]map[string]string, gn)
	for i := uint64(0); i < gn; i++ {
		ns, err := r.str()
		if err != nil {
			return nil, err
		}
		sn, err := r.uvarint()
		if err != nil {
			return nil, err
		}
		m := make(map[string]string, sn)
		for j := uint64(0); j < sn; j++ {
			sub, err := r.str()
			if err != nil {
				return nil, err
			}
			role, err := r.str()
			if err != nil {
				return nil, err
			}
			m[sub] = role
		}
		blob.Grants[ns] = m
	}

	un, err := r.uvarint()
	if err != nil {
		return nil, err
	}
	blob.Usage = make(map[string]quota.Usage, un)
	for i := uint64(0); i < un; i++ {
		ns, err := r.str()
		if err != nil {
			return nil, err
		}
		keys, err := r.i64()
		if err != nil {
			return nil, err
		}
		bytesN, err := r.i64()
		if err != nil {
			return nil, err
		}
		blob.Usage[ns] = quota.Usage{Keys: int(keys), Bytes: bytesN}
	}
	blob.Tally()
	return blob, nil
}
