package ns

import (
	"strings"
	"unicode/utf8"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
)

const (
	MaxNSLen  = 64
	MaxKeyLen = 192
	MaxOwner  = 64
)

// ValidNSID 允许小写字母开头，后接小写字母、数字、下划线、短横线。
func ValidNSID(id string) error {
	if id == "" || len(id) > MaxNSLen {
		return cherr.Wrap(cherr.ErrInvalidID, id)
	}
	if !utf8.ValidString(id) {
		return cherr.Wrap(cherr.ErrInvalidID, id)
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if i == 0 {
			if c < 'a' || c > 'z' {
				return cherr.Wrap(cherr.ErrInvalidID, id)
			}
			continue
		}
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			continue
		}
		return cherr.Wrap(cherr.ErrInvalidID, id)
	}
	return nil
}

// ValidKey 允许字母数字开头，中间可含 . _ / -；禁止 "."、".." 路径段与首尾斜杠。
func ValidKey(key string) error {
	if key == "" || len(key) > MaxKeyLen {
		return cherr.Wrap(cherr.ErrInvalidKey, key)
	}
	if !utf8.ValidString(key) || strings.ContainsRune(key, 0) {
		return cherr.Wrap(cherr.ErrInvalidKey, key)
	}
	if key[0] == '/' || key[len(key)-1] == '/' {
		return cherr.Wrap(cherr.ErrInvalidKey, key)
	}
	segStart := 0
	for i := 0; i <= len(key); i++ {
		if i < len(key) {
			c := key[i]
			ok := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
				c == '.' || c == '_' || c == '-' || c == '/'
			if !ok {
				return cherr.Wrap(cherr.ErrInvalidKey, key)
			}
			if c != '/' {
				continue
			}
		}
		seg := key[segStart:i]
		if seg == "" || seg == "." || seg == ".." {
			return cherr.Wrap(cherr.ErrInvalidKey, key)
		}
		if i == 0 {
			c := key[0]
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
				return cherr.Wrap(cherr.ErrInvalidKey, key)
			}
		}
		segStart = i + 1
	}
	c := key[0]
	if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
		return cherr.Wrap(cherr.ErrInvalidKey, key)
	}
	return nil
}

func ValidOwner(owner string) error {
	if owner == "" || len(owner) > MaxOwner {
		return cherr.Wrap(cherr.ErrInvalidArg, "owner")
	}
	if strings.ContainsRune(owner, 0) || !utf8.ValidString(owner) {
		return cherr.Wrap(cherr.ErrInvalidArg, "owner")
	}
	return nil
}

func ValidActor(actor string) error {
	if actor == "" {
		return nil
	}
	if len(actor) > MaxOwner || strings.ContainsRune(actor, 0) {
		return cherr.Wrap(cherr.ErrInvalidArg, "actor")
	}
	return nil
}
