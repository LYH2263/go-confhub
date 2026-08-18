package gray

import (
	"sort"
	"strings"

	cherr "github.com/LYH2263/go-confhub/internal/errors"
	"github.com/LYH2263/go-confhub/internal/meta"
)

type Rule = meta.GrayRule

func Normalize(g *Rule) (*Rule, error) {
	if g == nil {
		return nil, nil
	}
	if g.Percent < 0 || g.Percent > 100 {
		return nil, cherr.Wrap(cherr.ErrGrayInvalid, "percent")
	}
	out := meta.CloneGray(g)
	if out.AllowIDs != nil {
		seen := make(map[string]struct{}, len(out.AllowIDs))
		clean := make([]string, 0, len(out.AllowIDs))
		for _, id := range out.AllowIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			clean = append(clean, id)
		}
		sort.Strings(clean)
		out.AllowIDs = clean
	}
	if out.MatchTags != nil {
		clean := make(map[string]string, len(out.MatchTags))
		for k, v := range out.MatchTags {
			k = strings.TrimSpace(k)
			if k == "" {
				return nil, cherr.Wrap(cherr.ErrGrayInvalid, "empty tag key")
			}
			clean[k] = v
		}
		out.MatchTags = clean
	}
	if out.IsZero() {
		return nil, nil
	}
	return out, nil
}

func MustNormalize(g *Rule) *Rule {
	n, err := Normalize(g)
	if err != nil {
		return g
	}
	return n
}
