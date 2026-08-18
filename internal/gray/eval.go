package gray

import "github.com/LYH2263/go-confhub/internal/meta"

// Eval 评估客户端是否命中灰度。
// 规则：
//  1. nil / 全量发布 → 命中（读 Head）
//  2. AllowIDs 命中 → 命中（白名单优先）
//  3. MatchTags 若存在，必须全部匹配，否则未命中
//  4. Percent：Bucket(id) < percent；percent==0 且无白名单则未命中
func Eval(rule *Rule, c meta.ClientContext) bool {
	if rule == nil || rule.IsFullRollout() || rule.IsZero() {
		return true
	}
	if inAllow(rule.AllowIDs, c.ID) {
		return true
	}
	if len(rule.MatchTags) > 0 && !matchAllTags(rule.MatchTags, c.Tags) {
		return false
	}
	if rule.Percent <= 0 {
		return false
	}
	return InPercent(c.ID, rule.Percent)
}

func inAllow(ids []string, id string) bool {
	if id == "" || len(ids) == 0 {
		return false
	}
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func matchAllTags(need, have map[string]string) bool {
	if len(need) == 0 {
		return true
	}
	if have == nil {
		return false
	}
	for k, v := range need {
		if have[k] != v {
			return false
		}
	}
	return true
}

// Select 按灰度在 Head 与 Stable 之间选择版本指针。命中返回 Head，未命中返回 Stable。
// 返回的 hit 表示是否命中 Head 灰度；stable==0 且未命中时 ok=false。
func Select(e *meta.Entry, c meta.ClientContext) (rev int64, hit bool, ok bool) {
	if e == nil || e.Head == 0 {
		return 0, false, false
	}
	head := e.HeadMeta()
	if head == nil {
		return 0, false, false
	}
	if Eval(head.Gray, c) {
		return e.Head, true, true
	}
	return e.Head, false, true
}
