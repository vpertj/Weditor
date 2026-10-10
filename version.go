package main

import (
	"strconv"
	"strings"
)

// semver 极简实现，只覆盖 WEditor 实际发布的版本形式：
//
//	v1.2.3 / 1.2.3 / v1.2.3-dev-abc1234 / v1.0.0-dev
//
// 不做完整的 SemVer 2.0 预发布标识比较，原因见 compareSemver 注释。
type semver struct {
	major, minor, patch int
	prerelease          string
}

// parseSemver 解析版本号，容忍前导 v 与不完整的形式（v1 / v1.2）
func parseSemver(raw string) (semver, bool) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(strings.TrimPrefix(s, "v"), "V")
	if s == "" {
		return semver{}, false
	}

	var v semver
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		v.prerelease = s[i+1:]
		s = s[:i]
	}

	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return semver{}, false
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return semver{}, false
		}
		nums[i] = n
	}
	v.major, v.minor, v.patch = nums[0], nums[1], nums[2]
	return v, true
}

// baseVersion 返回去掉预发布标识的版本（1.0.0-dev-abc → 1.0.0）
func (v semver) baseVersion() semver {
	return semver{major: v.major, minor: v.minor, patch: v.patch}
}

// String 返回规范化的版本号（含预发布标识）
func (v semver) String() string {
	s := strconv.Itoa(v.major) + "." + strconv.Itoa(v.minor) + "." + strconv.Itoa(v.patch)
	if v.prerelease != "" {
		s += "-" + v.prerelease
	}
	return s
}

// isPrerelease 是否为预发布版本（CI 的 main 分支构建形如 v1.0.0-dev-<sha>）
func (v semver) isPrerelease() bool {
	return v.prerelease != ""
}

// compareSemver 比较两个版本，返回 -1 / 0 / 1。
//
// 规则：
//  1. 主次修订号按数值比较；
//  2. 数值相同且只有一个带预发布标识时，预发布版本更小（1.0.0-dev < 1.0.0）；
//  3. 两个都带预发布标识时视为相等 —— WEditor 的 dev 版本形如
//     v1.0.0-dev-<git sha>，sha 之间没有大小关系，按 SemVer 规则逐段比较
//     只会得到无意义的字典序结果。开发通道改用发布时间判定新旧，
//     见 pickDevRelease。
func compareSemver(a, b semver) int {
	if c := cmpInt(a.major, b.major); c != 0 {
		return c
	}
	if c := cmpInt(a.minor, b.minor); c != 0 {
		return c
	}
	if c := cmpInt(a.patch, b.patch); c != 0 {
		return c
	}
	switch {
	case a.prerelease == "" && b.prerelease != "":
		return 1
	case a.prerelease != "" && b.prerelease == "":
		return -1
	default:
		return 0
	}
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// newerThan 判断 candidate 是否比 current 新。
// 任一版本无法解析时返回 false —— 宁可漏报，不可误报触发一次破坏性更新。
func newerThan(candidate, current string) bool {
	cv, ok1 := parseSemver(candidate)
	cur, ok2 := parseSemver(current)
	if !ok1 || !ok2 {
		return false
	}
	return compareSemver(cv, cur) > 0
}
