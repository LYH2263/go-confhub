package sign

import "testing"

func TestHMACRoundTrip(t *testing.T) {
	k := NewKeyring()
	if err := k.PutHMAC("prod", "default", []byte("s3cret")); err != nil {
		t.Fatal(err)
	}
	payload := []byte("cfg")
	sig, err := k.Sign("prod", "k", "default", AlgoHMAC, 0, payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := k.Verify("prod", "k", "default", AlgoHMAC, 0, payload, sig); err != nil {
		t.Fatal(err)
	}
	if err := k.Verify("prod", "k", "default", AlgoHMAC, 0, payload, []byte("nope")); err == nil {
		t.Fatal("expected fail")
	}
}

func TestEd25519RoundTrip(t *testing.T) {
	pair, err := GenerateEd25519()
	if err != nil {
		t.Fatal(err)
	}
	k := NewKeyring()
	if err := k.PutEd("", "ed", pair); err != nil {
		t.Fatal(err)
	}
	msg := Canonical("n", "k", 2, []byte("p"))
	sig := EdSign(pair.Private, msg)
	if !EdVerify(pair.Public, msg, sig) {
		t.Fatal("ed verify")
	}
	if err := k.Verify("n", "k", "ed", AlgoEd25519, 2, []byte("p"), sig); err != nil {
		t.Fatal(err)
	}
}
