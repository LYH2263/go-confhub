package ns

import "testing"

func TestValidIDs(t *testing.T) {
	if err := ValidNSID("prod"); err != nil {
		t.Fatal(err)
	}
	if err := ValidNSID("Prod"); err == nil {
		t.Fatal("uppercase ns")
	}
	if err := ValidKey("app.timeout"); err != nil {
		t.Fatal(err)
	}
	if err := ValidKey("../etc/passwd"); err == nil {
		t.Fatal("path escape")
	}
	if err := ValidKey("/abs"); err == nil {
		t.Fatal("leading slash")
	}
}

func TestACLOwnerIsAdmin(t *testing.T) {
	r := NewRegistry(nil, nil)
	if _, err := r.Create("prod", "alice", 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := r.ACL().Check("prod", "alice", RoleAdmin); err != nil {
		t.Fatal(err)
	}
	if err := r.ACL().Check("prod", "bob", RoleReader); err == nil {
		t.Fatal("bob should be forbidden")
	}
	if err := r.ACL().Grant("prod", "bob", RoleReader); err != nil {
		t.Fatal(err)
	}
	if err := r.ACL().Check("prod", "bob", RoleReader); err != nil {
		t.Fatal(err)
	}
	if err := r.ACL().Check("prod", "bob", RoleWriter); err == nil {
		t.Fatal("reader cannot write")
	}
}
