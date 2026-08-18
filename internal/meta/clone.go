package meta

// CloneBytes 复制 payload / 签名，避免调用方改底层缓冲。
func CloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out
}

func CloneStringMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func CloneStrings(s []string) []string {
	if s == nil {
		return nil
	}
	out := make([]string, len(s))
	copy(out, s)
	return out
}

func CloneGray(g *GrayRule) *GrayRule {
	if g == nil {
		return nil
	}
	return &GrayRule{
		Percent:   g.Percent,
		AllowIDs:  CloneStrings(g.AllowIDs),
		MatchTags: CloneStringMap(g.MatchTags),
	}
}

func CloneVersion(v VersionMeta) VersionMeta {
	v.Payload = CloneBytes(v.Payload)
	v.Sig = CloneBytes(v.Sig)
	v.Gray = CloneGray(v.Gray)
	return v
}

func CloneEntry(e *Entry) *Entry {
	if e == nil {
		return nil
	}
	out := &Entry{
		NS:     e.NS,
		Key:    e.Key,
		Head:   e.Head,
		Stable: e.Stable,
	}
	if len(e.Versions) > 0 {
		out.Versions = make([]VersionMeta, len(e.Versions))
		for i, v := range e.Versions {
			out.Versions[i] = CloneVersion(v)
		}
	}
	return out
}

func CloneNS(n Namespace) Namespace {
	return n
}

func CloneClient(c ClientContext) ClientContext {
	return ClientContext{
		ID:   c.ID,
		Tags: CloneStringMap(c.Tags),
	}
}

func (e *Entry) Clone() *Entry { return CloneEntry(e) }

func (v VersionMeta) Clone() VersionMeta { return CloneVersion(v) }
