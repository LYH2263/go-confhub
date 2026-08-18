package ns

import (
	"sort"
	"sync"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
)

type Role int

const (
	RoleNone Role = iota
	RoleReader
	RoleWriter
	RoleAdmin
)

func (r Role) String() string {
	switch r {
	case RoleReader:
		return "reader"
	case RoleWriter:
		return "writer"
	case RoleAdmin:
		return "admin"
	default:
		return "none"
	}
}

func ParseRole(s string) (Role, error) {
	switch s {
	case "", "none":
		return RoleNone, nil
	case "reader", "read":
		return RoleReader, nil
	case "writer", "write":
		return RoleWriter, nil
	case "admin":
		return RoleAdmin, nil
	default:
		return RoleNone, cherr.Wrap(cherr.ErrInvalidArg, "role="+s)
	}
}

func (r Role) Covers(need Role) bool {
	if need == RoleNone {
		return true
	}
	return r >= need
}

type grantKey struct {
	ns      string
	subject string
}

// ACL 按命名空间授予主体角色；Owner 隐式 Admin。
type ACL struct {
	mu     sync.RWMutex
	grants map[grantKey]Role
	owners map[string]string
}

func NewACL() *ACL {
	return &ACL{
		grants: make(map[grantKey]Role),
		owners: make(map[string]string),
	}
}

func (a *ACL) SetOwner(nsID, owner string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.owners[nsID] = owner
}

func (a *ACL) RemoveNS(nsID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.owners, nsID)
	for k := range a.grants {
		if k.ns == nsID {
			delete(a.grants, k)
		}
	}
}

func (a *ACL) Grant(nsID, subject string, role Role) error {
	if err := ValidNSID(nsID); err != nil {
		return err
	}
	if err := ValidActor(subject); err != nil || subject == "" {
		return cherr.Wrap(cherr.ErrInvalidArg, "subject")
	}
	if role == RoleNone {
		a.mu.Lock()
		delete(a.grants, grantKey{ns: nsID, subject: subject})
		a.mu.Unlock()
		return nil
	}
	a.mu.Lock()
	a.grants[grantKey{ns: nsID, subject: subject}] = role
	a.mu.Unlock()
	return nil
}

func (a *ACL) RoleOf(nsID, subject string) Role {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if subject != "" && a.owners[nsID] == subject {
		return RoleAdmin
	}
	if subject == "" {
		return RoleNone
	}
	return a.grants[grantKey{ns: nsID, subject: subject}]
}

func (a *ACL) Check(nsID, subject string, need Role) error {
	if need == RoleNone {
		return nil
	}
	got := a.RoleOf(nsID, subject)
	if got.Covers(need) {
		return nil
	}
	return cherr.Wrap(cherr.ErrForbidden, nsID+"/"+subject+" need="+need.String())
}

type Grant struct {
	Subject string `json:"subject"`
	Role    string `json:"role"`
}

func (a *ACL) List(nsID string) []Grant {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]Grant, 0, 8)
	if owner := a.owners[nsID]; owner != "" {
		out = append(out, Grant{Subject: owner, Role: RoleAdmin.String()})
	}
	for k, r := range a.grants {
		if k.ns != nsID {
			continue
		}
		if a.owners[nsID] == k.subject {
			continue
		}
		out = append(out, Grant{Subject: k.subject, Role: r.String()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Subject < out[j].Subject })
	return out
}

func (a *ACL) Export() map[string]map[string]string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make(map[string]map[string]string)
	for k, r := range a.grants {
		m := out[k.ns]
		if m == nil {
			m = make(map[string]string)
			out[k.ns] = m
		}
		m[k.subject] = r.String()
	}
	return out
}

func (a *ACL) Import(m map[string]map[string]string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.grants = make(map[grantKey]Role)
	for nsID, subs := range m {
		for sub, rs := range subs {
			role, err := ParseRole(rs)
			if err != nil || role == RoleNone {
				continue
			}
			a.grants[grantKey{ns: nsID, subject: sub}] = role
		}
	}
}
