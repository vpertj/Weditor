package main

import "testing"

func TestParseSemver(t *testing.T) {
	cases := []struct {
		in      string
		want    semver
		wantErr bool
	}{
		{"1.0.0", semver{1, 0, 0, ""}, false},
		{"v1.0.0", semver{1, 0, 0, ""}, false},
		{"V2.3.4", semver{2, 3, 4, ""}, false},
		{"v1.0.0-dev-07740b4", semver{1, 0, 0, "dev-07740b4"}, false},
		{"v1.2.3+build.5", semver{1, 2, 3, "build.5"}, false},
		{"v1", semver{1, 0, 0, ""}, false},
		{"v1.2", semver{1, 2, 0, ""}, false},
		{"", semver{}, true},
		{"v", semver{}, true},
		{"vx.y.z", semver{}, true},
		{"v1.2.3.4", semver{}, true},
	}
	for _, c := range cases {
		got, ok := parseSemver(c.in)
		if c.wantErr {
			if !ok {
				continue
			}
			t.Errorf("parseSemver(%q) 期望失败，实际得到 %+v", c.in, got)
			continue
		}
		if !ok {
			t.Errorf("parseSemver(%q) 意外失败", c.in)
			continue
		}
		if got != c.want {
			t.Errorf("parseSemver(%q) = %+v, 期望 %+v", c.in, got, c.want)
		}
	}
}

func TestSemverBaseVersion(t *testing.T) {
	v, _ := parseSemver("v1.2.3-dev-abc")
	if got := v.baseVersion().String(); got != "1.2.3" {
		t.Errorf("baseVersion() = %q, 期望 1.2.3", got)
	}
}

func TestCompareSemver(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.1", "1.0.0", 1},
		{"1.1.0", "1.0.9", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.0.0", "1.0.1", -1},
		// 正式版大于同号预发布版
		{"1.0.0", "1.0.0-dev-abc", 1},
		{"1.0.0-dev-abc", "1.0.0", -1},
		// 两个 dev 版本无法定序，视为相等（由发布时间判定）
		{"1.0.0-dev-aaa", "1.0.0-dev-zzz", 0},
		// 基础版本更高的 dev 仍然更新
		{"1.1.0-dev-abc", "1.0.0", 1},
	}
	for _, c := range cases {
		av, okA := parseSemver(c.a)
		bv, okB := parseSemver(c.b)
		if !okA || !okB {
			t.Fatalf("解析失败: %q / %q", c.a, c.b)
		}
		if got := compareSemver(av, bv); got != c.want {
			t.Errorf("compareSemver(%q, %q) = %d, 期望 %d", c.a, c.b, got, c.want)
		}
	}
}

func TestNewerThan(t *testing.T) {
	cases := []struct {
		candidate, current string
		want               bool
	}{
		{"v1.1.0", "v1.0.0", true},
		{"v1.0.0", "v1.0.0", false},
		{"v1.0.0", "v1.1.0", false},
		{"v1.0.0", "v1.0.0-dev-abc", true},
		// 无法解析时一律不更新，避免误判触发破坏性更新
		{"garbage", "v1.0.0", false},
		{"v1.1.0", "", false},
	}
	for _, c := range cases {
		if got := newerThan(c.candidate, c.current); got != c.want {
			t.Errorf("newerThan(%q, %q) = %v, 期望 %v", c.candidate, c.current, got, c.want)
		}
	}
}
