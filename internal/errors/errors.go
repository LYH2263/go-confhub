// Package errors 定义配置中心的哨兵错误，供库面与 HTTP 层统一判断。
package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("confhub: not found")
	ErrAlreadyExists = errors.New("confhub: already exists")
	ErrInvalidID     = errors.New("confhub: invalid id")
	ErrInvalidKey    = errors.New("confhub: invalid key")
	ErrInvalidArg    = errors.New("confhub: invalid argument")
	ErrForbidden     = errors.New("confhub: forbidden")
	ErrUnauthorized  = errors.New("confhub: unauthorized")
	ErrConflict      = errors.New("confhub: conflict")
	ErrSignFailed    = errors.New("confhub: signature verification failed")
	ErrSignRequired  = errors.New("confhub: signature required")
	ErrUnknownAlgo   = errors.New("confhub: unknown signature algorithm")
	ErrUnknownKey    = errors.New("confhub: unknown signing key")
	ErrQuotaKeys     = errors.New("confhub: namespace key quota exceeded")
	ErrQuotaBytes    = errors.New("confhub: namespace byte quota exceeded")
	ErrNoStable      = errors.New("confhub: no stable version for gray miss")
	ErrNotPublished  = errors.New("confhub: version is not published")
	ErrBadRevision   = errors.New("confhub: revision out of range")
	ErrStaleWatch    = errors.New("confhub: watch subscription closed")
	ErrNamespaceGone = errors.New("confhub: namespace deleted")
	ErrPayloadEmpty  = errors.New("confhub: payload must not be empty")
	ErrGrayInvalid   = errors.New("confhub: invalid gray rule")
	ErrSnapshot      = errors.New("confhub: snapshot codec error")
	ErrChecksum      = errors.New("confhub: snapshot checksum mismatch")
	ErrClosed        = errors.New("confhub: hub is closed")
)

type withRef struct {
	err error
	ref string
}

func (w withRef) Error() string {
	if w.ref == "" {
		return w.err.Error()
	}
	return w.err.Error() + ": " + w.ref
}

func (w withRef) Unwrap() error { return w.err }

// Wrap 给哨兵错误附加资源定位信息（ns/key/rev）。
func Wrap(err error, ref string) error {
	if err == nil {
		return nil
	}
	return withRef{err: err, ref: ref}
}

func Is(err, target error) bool { return errors.Is(err, target) }

func New(msg string) error { return errors.New(msg) }

func Fmt(format string, args ...any) error { return fmt.Errorf(format, args...) }
