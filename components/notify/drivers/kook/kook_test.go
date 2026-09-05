package kook

import (
	"context"
	"testing"
	_ "time/tzdata" // 内嵌 zoneinfo，保证无系统 tzdata 的环境下测试同样可跑
)

func TestCfgLocation(t *testing.T) {
	tests := []struct {
		name string
		tz   string
		want string // 期望的 time.Location.String()，Local 表示服务器本地时区
	}{
		{
			name: "empty_timezone_falls_back_to_local",
			tz:   "",
			want: "Local",
		},
		{
			name: "valid_iana_name",
			tz:   "Asia/Shanghai",
			want: "Asia/Shanghai",
		},
		{
			name: "invalid_timezone_falls_back_to_local",
			tz:   "Foo/Bar",
			want: "Local",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfgLocation(context.Background(), tt.tz)
			if got.String() != tt.want {
				t.Errorf("cfgLocation(%q) = %v, want %v", tt.tz, got, tt.want)
			}
		})
	}
}
