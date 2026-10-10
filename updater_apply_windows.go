//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// applyUpdatePlatform 就地替换 exe。
//
// Windows 允许重命名正在运行的 exe（FILE_SHARE_DELETE），但不允许覆盖或删除它。
// 因此先把当前 exe 改名让出文件名，再把新版本放到原路径，最后拉起新进程。
func (a *App) applyUpdatePlatform(stagedPath, version string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法定位程序路径：%v", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	if looksLikeDevBuild(exe) {
		return errors.New("当前是开发构建，已跳过自动更新（请手动安装新版本）")
	}

	dir := filepath.Dir(exe)
	if err := ensureWritable(dir); err != nil {
		return fmt.Errorf("没有权限修改 %s，请手动安装新版本", dir)
	}

	stale := exe + ".old"
	// 上一次更新留下的旧文件，那时占用它的进程早已退出
	os.Remove(stale)

	if err := os.Rename(exe, stale); err != nil {
		return fmt.Errorf("替换失败（文件可能被占用）：%v，请手动安装新版本", err)
	}
	if err := copyFile(stagedPath, exe); err != nil {
		// 回滚，保证用户手里的原程序仍然可启动
		os.Rename(stale, exe)
		return fmt.Errorf("写入新版本失败：%v", err)
	}

	// 先拉起新版本再退出，避免出现"旧进程已退、新进程没起来"的空窗
	cmd := exec.Command(exe)
	cmd.Dir = dir
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动新版本失败：%v", err)
	}
	return nil
}

// cleanupPlatformArtifacts 删除上次更新留下的 WEditor.exe.old。
// 旧进程退出前无法删除它，所以只能等到下一次启动。
func cleanupPlatformArtifacts() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	os.Remove(exe + ".old")
}
