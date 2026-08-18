// Package meta 存放跨包共享的配置中心数据结构。
package meta

import "time"

const (
	KindPublish  = "publish"
	KindRollback = "rollback"
	KindDelete   = "delete"
	KindCreateNS = "create_ns"
	KindDeleteNS = "delete_ns"
	KindImport   = "import"
)

// Namespace 对应设计文档 Namespace{ID, Owner, CreatedAt, MaxKeys, MaxBytes}。
type Namespace struct {
	ID        string    `json:"id"`
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
	MaxKeys   int       `json:"max_keys"`
	MaxBytes  int64     `json:"max_bytes"`
}

// GrayRule 灰度：白名单优先，其次标签全匹配，再按百分比分桶。
type GrayRule struct {
	Percent   int               `json:"percent"`
	AllowIDs  []string          `json:"allow_ids,omitempty"`
	MatchTags map[string]string `json:"match_tags,omitempty"`
}

func (g *GrayRule) IsZero() bool {
	if g == nil {
		return true
	}
	return g.Percent <= 0 && len(g.AllowIDs) == 0 && len(g.MatchTags) == 0
}

func (g *GrayRule) IsFullRollout() bool {
	if g == nil {
		return true
	}
	return g.Percent >= 100 && len(g.AllowIDs) == 0 && len(g.MatchTags) == 0
}

// VersionMeta 一条不可变版本记录。
type VersionMeta struct {
	Rev     int64     `json:"rev"`
	Payload []byte    `json:"payload"`
	Sig     []byte    `json:"sig,omitempty"`
	Algo    string    `json:"algo,omitempty"`
	Author  string    `json:"author,omitempty"`
	Ts      time.Time `json:"ts"`
	Gray    *GrayRule `json:"gray,omitempty"`
}

// Entry 同一 (ns, key) 的版本链与两个指针：Head（最新）与 Stable（灰度未命中回落）。
type Entry struct {
	NS       string        `json:"ns"`
	Key      string        `json:"key"`
	Versions []VersionMeta `json:"versions"`
	Head     int64         `json:"head"`
	Stable   int64         `json:"stable"`
}

func (e *Entry) Len() int {
	if e == nil {
		return 0
	}
	return len(e.Versions)
}

func (e *Entry) ByRev(rev int64) *VersionMeta {
	if e == nil || rev <= 0 || int(rev) > len(e.Versions) {
		return nil
	}
	v := &e.Versions[rev-1]
	if v.Rev != rev {
		return nil
	}
	return v
}

func (e *Entry) HeadMeta() *VersionMeta {
	if e == nil {
		return nil
	}
	return e.ByRev(e.Head)
}

func (e *Entry) StableMeta() *VersionMeta {
	if e == nil {
		return nil
	}
	return e.ByRev(e.Stable)
}

// ClientContext 客户端身份，用于灰度评估。
type ClientContext struct {
	ID   string            `json:"id"`
	Tags map[string]string `json:"tags,omitempty"`
}

// WatchEvent 订阅投递。
type WatchEvent struct {
	NS   string `json:"ns"`
	Key  string `json:"key"`
	Rev  int64  `json:"rev"`
	Kind string `json:"kind"`
}

// AuditRecord 变更审计。
type AuditRecord struct {
	Seq    int64     `json:"seq"`
	Ts     time.Time `json:"ts"`
	NS     string    `json:"ns"`
	Key    string    `json:"key,omitempty"`
	Rev    int64     `json:"rev,omitempty"`
	Kind   string    `json:"kind"`
	Actor  string    `json:"actor,omitempty"`
	Detail string    `json:"detail,omitempty"`
	Bytes  int       `json:"bytes,omitempty"`
	OK     bool      `json:"ok"`
	Err    string    `json:"err,omitempty"`
}

// SnapshotHeader 快照元信息。
type SnapshotHeader struct {
	Magic     string    `json:"magic"`
	Format    int       `json:"format"`
	CreatedAt time.Time `json:"created_at"`
	NSCount   int       `json:"ns_count"`
	KeyCount  int       `json:"key_count"`
	VerCount  int       `json:"ver_count"`
}
