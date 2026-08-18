// Package version 负责版本号算术、格式化与「不可跳号」检查。
package version

import (
	"strconv"
	"strings"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
)

func Next(cur int64) int64 {
	if cur < 0 {
		return 1
	}
	return cur + 1
}

func Format(rev int64) string {
	return "r" + strconv.FormatInt(rev, 10)
}

func Parse(s string) (int64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "r")
	s = strings.TrimPrefix(s, "R")
	if s == "" {
		return 0, cherr.Wrap(cherr.ErrBadRevision, "empty")
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n <= 0 {
		return 0, cherr.Wrap(cherr.ErrBadRevision, s)
	}
	return n, nil
}

func Cmp(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// CheckStrictInc 断言 versions[i].Rev == i+1，读最新不得跳号。
func CheckStrictInc(revs []int64) error {
	for i, r := range revs {
		want := int64(i + 1)
		if r != want {
			return cherr.Wrap(cherr.ErrBadRevision, "index "+strconv.Itoa(i)+" has "+Format(r)+" want "+Format(want))
		}
	}
	return nil
}

func InRange(rev, lo, hi int64) bool {
	return rev >= lo && rev <= hi
}

func Max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
