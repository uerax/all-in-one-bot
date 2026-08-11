package router

import (
	"testing"
)

func TestAdminIDsSet(t *testing.T) {
	if AdminIDsSet(nil) != nil {
		t.Error("nil 切片应返回 nil")
	}
	if AdminIDsSet([]int64{}) != nil {
		t.Error("空切片应返回 nil")
	}
	set := AdminIDsSet([]int64{1, 2, 2, 3})
	if len(set) != 3 || !set[1] || !set[2] || !set[3] {
		t.Errorf("AdminIDsSet 结果不正确: %v", set)
	}
}

func TestIsAuthorized(t *testing.T) {
	admin := AdminIDsSet([]int64{100})
	cases := []struct {
		name     string
		adminIDs map[int64]bool
		cmd      string
		senderID int64
		want     bool
	}{
		{"未配置白名单 → 全部放行", nil, "/crocodile_list", 1, true},
		{"/chatid 放行（sender 非管理员）", admin, "/chatid", 1, true},
		{"管理员放行", admin, "/crocodile_list", 100, true},
		{"非管理员丢弃", admin, "/crocodile_list", 200, false},
		{"sender 0（未知）丢弃", admin, "/crocodile_list", 0, false},
		{"未配置时 /chatid 也放行", nil, "/chatid", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isAuthorized(tc.adminIDs, tc.cmd, tc.senderID)
			if got != tc.want {
				t.Errorf("isAuthorized(%v, %q, %d) = %v, want %v",
					tc.adminIDs, tc.cmd, tc.senderID, got, tc.want)
			}
		})
	}
}
