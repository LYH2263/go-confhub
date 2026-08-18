package sign

import (
	"crypto/ed25519"
	"sync"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
)

type hmacKey struct {
	secret []byte
}

type edKey struct {
	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
}

type nsKeyID struct {
	ns  string
	id  string
	alg string
}

// Keyring 按命名空间+KeyID 存放 HMAC 密钥与 Ed25519 公钥（可选私钥用于测试签发）。
type Keyring struct {
	mu    sync.RWMutex
	hmac  map[nsKeyID]hmacKey
	ed    map[nsKeyID]edKey
	deflt string
}

func NewKeyring() *Keyring {
	return &Keyring{
		hmac:  make(map[nsKeyID]hmacKey),
		ed:    make(map[nsKeyID]edKey),
		deflt: "default",
	}
}

func (k *Keyring) DefaultID() string { return k.deflt }

func (k *Keyring) SetDefaultID(id string) {
	if id != "" {
		k.deflt = id
	}
}

func (k *Keyring) PutHMAC(ns, keyID string, secret []byte) error {
	if len(secret) == 0 {
		return cherr.Wrap(cherr.ErrInvalidArg, "empty hmac secret")
	}
	if keyID == "" {
		keyID = k.deflt
	}
	cp := make([]byte, len(secret))
	copy(cp, secret)
	k.mu.Lock()
	k.hmac[nsKeyID{ns: ns, id: keyID, alg: AlgoHMAC}] = hmacKey{secret: cp}
	k.mu.Unlock()
	return nil
}

func (k *Keyring) PutEd(ns, keyID string, pair EdKeyPair) error {
	if len(pair.Public) != ed25519.PublicKeySize {
		return cherr.Wrap(cherr.ErrInvalidArg, "ed public")
	}
	if keyID == "" {
		keyID = k.deflt
	}
	k.mu.Lock()
	k.ed[nsKeyID{ns: ns, id: keyID, alg: AlgoEd25519}] = edKey{
		pub:  append(ed25519.PublicKey(nil), pair.Public...),
		priv: append(ed25519.PrivateKey(nil), pair.Private...),
	}
	k.mu.Unlock()
	return nil
}

func (k *Keyring) lookupHMAC(ns, keyID string) ([]byte, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if keyID == "" {
		keyID = k.deflt
	}
	if v, ok := k.hmac[nsKeyID{ns: ns, id: keyID, alg: AlgoHMAC}]; ok {
		return v.secret, true
	}
	if v, ok := k.hmac[nsKeyID{ns: "", id: keyID, alg: AlgoHMAC}]; ok {
		return v.secret, true
	}
	return nil, false
}

func (k *Keyring) lookupEd(ns, keyID string) (edKey, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if keyID == "" {
		keyID = k.deflt
	}
	if v, ok := k.ed[nsKeyID{ns: ns, id: keyID, alg: AlgoEd25519}]; ok {
		return v, true
	}
	if v, ok := k.ed[nsKeyID{ns: "", id: keyID, alg: AlgoEd25519}]; ok {
		return v, true
	}
	return edKey{}, false
}

func (k *Keyring) Sign(ns, key, keyID, algo string, rev int64, payload []byte) ([]byte, error) {
	algo = NormalizeAlgo(algo)
	msg := Canonical(ns, key, rev, payload)
	switch algo {
	case AlgoHMAC, "":
		secret, ok := k.lookupHMAC(ns, keyID)
		if !ok {
			return nil, cherr.Wrap(cherr.ErrUnknownKey, keyID)
		}
		return HMACSign(secret, msg), nil
	case AlgoEd25519:
		ek, ok := k.lookupEd(ns, keyID)
		if !ok || len(ek.priv) != ed25519.PrivateKeySize {
			return nil, cherr.Wrap(cherr.ErrUnknownKey, keyID)
		}
		return EdSign(ek.priv, msg), nil
	default:
		return nil, cherr.Wrap(cherr.ErrUnknownAlgo, algo)
	}
}

func (k *Keyring) Verify(ns, key, keyID, algo string, rev int64, payload, sig []byte) error {
	algo = NormalizeAlgo(algo)
	if algo == "" {
		algo = AlgoHMAC
	}
	if len(sig) == 0 {
		return cherr.ErrSignFailed
	}
	msg := Canonical(ns, key, rev, payload)
	switch algo {
	case AlgoHMAC:
		secret, ok := k.lookupHMAC(ns, keyID)
		if !ok {
			return cherr.Wrap(cherr.ErrUnknownKey, keyID)
		}
		if !HMACVerify(secret, msg, sig) {
			return cherr.ErrSignFailed
		}
		return nil
	case AlgoEd25519:
		ek, ok := k.lookupEd(ns, keyID)
		if !ok {
			return cherr.Wrap(cherr.ErrUnknownKey, keyID)
		}
		if !EdVerify(ek.pub, msg, sig) {
			return cherr.ErrSignFailed
		}
		return nil
	default:
		return cherr.Wrap(cherr.ErrUnknownAlgo, algo)
	}
}

func (k *Keyring) HasAny() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return len(k.hmac) > 0 || len(k.ed) > 0
}
