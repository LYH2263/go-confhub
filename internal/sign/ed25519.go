package sign

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
)

type EdKeyPair struct {
	Public  ed25519.PublicKey
	Private ed25519.PrivateKey
}

func GenerateEd25519() (EdKeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return EdKeyPair{}, err
	}
	return EdKeyPair{Public: pub, Private: priv}, nil
}

func EdSign(priv ed25519.PrivateKey, message []byte) []byte {
	if len(priv) != ed25519.PrivateKeySize {
		return nil
	}
	return ed25519.Sign(priv, message)
}

func EdVerify(pub ed25519.PublicKey, message, sig []byte) bool {
	if len(pub) != ed25519.PublicKeySize || len(sig) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(pub, message, sig)
}

func ParseEdPublic(hexPub string) (ed25519.PublicKey, error) {
	b, err := hex.DecodeString(hexPub)
	if err != nil {
		return nil, cherr.Wrap(cherr.ErrInvalidArg, "ed25519 public")
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, cherr.Wrap(cherr.ErrInvalidArg, "ed25519 public length")
	}
	return ed25519.PublicKey(b), nil
}

func ParseEdPrivate(hexPriv string) (ed25519.PrivateKey, error) {
	b, err := hex.DecodeString(hexPriv)
	if err != nil {
		return nil, cherr.Wrap(cherr.ErrInvalidArg, "ed25519 private")
	}
	if len(b) != ed25519.PrivateKeySize {
		return nil, cherr.Wrap(cherr.ErrInvalidArg, "ed25519 private length")
	}
	return ed25519.PrivateKey(b), nil
}

func EncodeEdPublic(pub ed25519.PublicKey) string {
	return hex.EncodeToString(pub)
}

func EncodeEdPrivate(priv ed25519.PrivateKey) string {
	return hex.EncodeToString(priv)
}
