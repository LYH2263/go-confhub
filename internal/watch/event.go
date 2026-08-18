package watch

import "github.com/LYH2263/go-confhub/internal/meta"

type Event = meta.WatchEvent

const (
	KindPublish  = meta.KindPublish
	KindRollback = meta.KindRollback
	KindDelete   = meta.KindDelete
)

func Match(filterNS, filterKey string, ev Event) bool {
	if filterNS != "" && filterNS != ev.NS {
		return false
	}
	if filterKey != "" && filterKey != ev.Key {
		return false
	}
	return true
}

func NewEvent(ns, key, kind string, rev int64) Event {
	return Event{NS: ns, Key: key, Rev: rev, Kind: kind}
}
