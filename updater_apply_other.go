//go:build !darwin && !windows

package main

import "errors"

// applyUpdatePlatform Linux 等平台暂不支持自动安装。
// 更新器仍会正常检测并下载，前端据此降级为"打开下载页 + 手动安装"。
func (a *App) applyUpdatePlatform(stagedPath, version string) error {
	return errors.New("当前平台暂不支持自动更新，请手动安装新版本")
}

func cleanupPlatformArtifacts() {}
