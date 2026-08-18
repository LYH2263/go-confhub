package confhub

import (
	"time"

	"github.com/LYH2263/go-confhub/internal/meta"
)

type Namespace = meta.Namespace
type GrayRule = meta.GrayRule
type VersionMeta = meta.VersionMeta
type Entry = meta.Entry
type ClientContext = meta.ClientContext
type WatchEvent = meta.WatchEvent
type AuditRecord = meta.AuditRecord

// Value 是 Get 的返回：按灰度选定的不可变版本。
type Value struct {
	NS      string    `json:"ns"`
	Key     string    `json:"key"`
	Rev     int64     `json:"rev"`
	Payload []byte    `json:"payload"`
	Author  string    `json:"author,omitempty"`
	Ts      time.Time `json:"ts"`
	Gray    *GrayRule `json:"gray,omitempty"`
	Algo    string    `json:"algo,omitempty"`
	Head    int64     `json:"head"`
	Stable  int64     `json:"stable"`
	Hit     bool      `json:"hit"`
}

func valueFrom(e *Entry, v *VersionMeta, hit bool) *Value {
	if e == nil || v == nil {
		return nil
	}
	return &Value{
		NS:      e.NS,
		Key:     e.Key,
		Rev:     v.Rev,
		Payload: meta.CloneBytes(v.Payload),
		Author:  v.Author,
		Ts:      v.Ts,
		Gray:    meta.CloneGray(v.Gray),
		Algo:    v.Algo,
		Head:    e.Head,
		Stable:  e.Stable,
		Hit:     hit,
	}
}

type PutOptions struct {
	Author    string
	Actor     string
	Algo      string
	Signature []byte
	KeyID     string
	Gray      *GrayRule
	// BindRev>0 时签名绑定该版本号，且必须等于即将写入的 next rev。
	BindRev int64
}

type NSSpec struct {
	ID       string
	Owner    string
	MaxKeys  int
	MaxBytes int64
	Actor    string
}
