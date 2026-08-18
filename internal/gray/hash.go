package gray

import (
	"hash/fnv"
)

// Bucket 将客户端 ID 稳定映射到 [0, 99]，用于百分比灰度。
func Bucket(clientID string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(clientID))
	return int(h.Sum64() % 100)
}

func InPercent(clientID string, percent int) bool {
	if percent <= 0 {
		return false
	}
	if percent >= 100 {
		return true
	}
	return Bucket(clientID) < percent
}

func Hash64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

// CombineSeed 把 ns/key/client 混在一起时仍保持与 Bucket(client) 独立；
// 当前百分比灰度只按 clientID，此函数给需要「按键打散」的调用方使用。
func CombineSeed(ns, key, clientID string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(ns))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(key))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(clientID))
	return int(h.Sum64() % 100)
}
