package store

import (
	"sort"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/meta"
	"github.com/LYH2263/go-confhub/internal/version"
)

func nsKey(ns, key string) string { return ns + "\x00" + key }

func nextRev(e *meta.Entry) int64 {
	if e == nil || len(e.Versions) == 0 {
		return 1
	}
	return version.Next(e.Versions[len(e.Versions)-1].Rev)
}

// AppendVersion 在条目末尾追加一个版本，版本号必须是 len+1。
func AppendVersion(e *meta.Entry, v meta.VersionMeta) (*meta.Entry, error) {
	if e == nil {
		return nil, cherr.ErrInvalidArg
	}
	want := nextRev(e)
	if v.Rev == 0 {
		v.Rev = want
	}
	if v.Rev != want {
		return nil, cherr.Wrap(cherr.ErrBadRevision, version.Format(v.Rev)+" want="+version.Format(want))
	}
	e.Versions = append(e.Versions, v)
	e.Head = v.Rev
	if v.Gray == nil || v.Gray.IsFullRollout() || v.Gray.IsZero() {
		e.Stable = v.Rev
	} else if e.Stable == 0 {
		// 首版即灰度：尚未形成稳定指针；未命中客户端读不到已发布稳定版。
		e.Stable = 0
	}
	return e, nil
}

func NewEntry(nsName, key string, v meta.VersionMeta) (*meta.Entry, error) {
	e := &meta.Entry{NS: nsName, Key: key}
	if v.Rev == 0 {
		v.Rev = 1
	}
	if v.Rev != 1 {
		return nil, cherr.Wrap(cherr.ErrBadRevision, version.Format(v.Rev))
	}
	e.Versions = []meta.VersionMeta{v}
	e.Head = 1
	if v.Gray == nil || v.Gray.IsFullRollout() || v.Gray.IsZero() {
		e.Stable = 1
	}
	return e, nil
}

func RollbackHead(e *meta.Entry, rev int64) error {
	if e == nil {
		return cherr.ErrNotFound
	}
	v := e.ByRev(rev)
	if v == nil {
		return cherr.Wrap(cherr.ErrBadRevision, version.Format(rev))
	}
	e.Head = rev
	if v.Gray == nil || v.Gray.IsFullRollout() || v.Gray.IsZero() {
		e.Stable = rev
		return nil
	}
	// 回滚到灰度版：Stable 取 <= rev 的最近全量发布。
	e.Stable = lastStableAtOrBefore(e, rev)
	return nil
}

func lastStableAtOrBefore(e *meta.Entry, rev int64) int64 {
	for i := int(rev); i >= 1; i-- {
		v := e.ByRev(int64(i))
		if v == nil {
			continue
		}
		if v.Gray == nil || v.Gray.IsFullRollout() || v.Gray.IsZero() {
			return v.Rev
		}
	}
	return 0
}

func BytesOf(e *meta.Entry) int64 {
	if e == nil {
		return 0
	}
	var n int64
	for i := range e.Versions {
		n += int64(len(e.Versions[i].Payload))
	}
	return n
}

func KeysSorted(entries map[string]*meta.Entry, nsName string) []string {
	out := make([]string, 0, 8)
	for _, e := range entries {
		if e.NS == nsName {
			out = append(out, e.Key)
		}
	}
	sort.Strings(out)
	return out
}

func ValidateChain(e *meta.Entry) error {
	if e == nil {
		return cherr.ErrNotFound
	}
	if len(e.Versions) == 0 {
		return cherr.Wrap(cherr.ErrBadRevision, e.NS+"/"+e.Key+" empty")
	}
	for i, v := range e.Versions {
		want := int64(i + 1)
		if v.Rev != want {
			return cherr.Wrap(cherr.ErrBadRevision, e.NS+"/"+e.Key+" gap")
		}
	}
	if e.Head < 1 || int(e.Head) > len(e.Versions) {
		return cherr.Wrap(cherr.ErrBadRevision, "head")
	}
	if e.Stable < 0 || int(e.Stable) > len(e.Versions) {
		return cherr.Wrap(cherr.ErrBadRevision, "stable")
	}
	return nil
}
