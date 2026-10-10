//go:build darwin

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// applyScript 由独立脚本完成替换。
//
// 为什么不在进程内替换：把新应用包搬进原位后本进程仍持有旧 inode，
// 且 App 退出后再无执行机会。脚本先等本进程退出，再做替换和重启，
// 这样替换始终发生在"没有进程在用这个包"的时刻。
//
// 参数：1=PID 2=当前应用包 3=新应用包 4=临时目录
const applyScript = `#!/bin/sh
set -u

pid="$1"
target="$2"
new="$3"
staging="$4"

# 等应用完全退出，否则替换掉的包仍被旧进程占用
while kill -0 "$pid" 2>/dev/null; do
  sleep 0.3
done
sleep 0.5

old="$staging/previous.app"

if [ -e "$target" ]; then
  # 旧包先挪进临时目录（同分区，纯 rename），失败才强删
  mv "$target" "$old" 2>/dev/null || rm -rf "$target"
fi

ok=1
if mv "$new" "$target" 2>/dev/null; then
  touch "$target" 2>/dev/null || true
  # 自己下载的包理论上不带隔离属性，这里兜底清一次
  xattr -dr com.apple.quarantine "$target" 2>/dev/null || true
  open -n "$target"
  ok=0
fi

if [ "$ok" -ne 0 ]; then
  # 放新包失败：把旧包还原回去，至少让用户还有可用的版本
  if [ -e "$old" ] && [ ! -e "$target" ]; then
    mv "$old" "$target" 2>/dev/null || true
    open -n "$target" 2>/dev/null || true
  fi
fi

# 清理交给子进程延迟执行：脚本不能在自己还在运行时删掉自己
( sleep 3; rm -rf "$staging" ) >/dev/null 2>&1 &
`

// applyUpdatePlatform 用解包出的新 .app 替换当前应用包，替换完成后由脚本重启应用
func (a *App) applyUpdatePlatform(stagedPath, version string) error {
	bundlePath, err := currentBundlePath()
	if err != nil {
		return err
	}
	if looksLikeDevBuild(filepath.Join(bundlePath, "Contents", "MacOS", APP_NAME)) {
		return errors.New("当前是开发构建，已跳过自动更新（请手动安装新版本）")
	}

	parentDir := filepath.Dir(bundlePath)
	if err := ensureWritable(parentDir); err != nil {
		return fmt.Errorf("没有权限修改 %s，请手动安装新版本", parentDir)
	}

	// 临时目录放在应用包同级：最后的 mv 是同分区 rename，原子且不丢元数据
	staging, err := os.MkdirTemp(parentDir, ".weditor-update-")
	if err != nil {
		return fmt.Errorf("创建临时目录失败：%v", err)
	}

	// ditto 比 Go 的 archive/zip 更能原样保留 .app 内的权限位和符号链接
	if out, err := exec.Command("ditto", "-x", "-k", stagedPath, staging).CombinedOutput(); err != nil {
		os.RemoveAll(staging)
		return fmt.Errorf("解压安装包失败：%v %s", err, strings.TrimSpace(string(out)))
	}

	newBundle, err := findAppBundle(staging)
	if err != nil {
		os.RemoveAll(staging)
		return err
	}

	script := filepath.Join(staging, "apply.sh")
	if err := os.WriteFile(script, []byte(applyScript), 0755); err != nil {
		os.RemoveAll(staging)
		return fmt.Errorf("写入更新脚本失败：%v", err)
	}

	cmd := exec.Command("/bin/sh", script, strconv.Itoa(os.Getpid()), bundlePath, newBundle, staging)
	// Setsid 让脚本脱离本进程所在的会话：应用退出后脚本仍能继续执行
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		os.RemoveAll(staging)
		return fmt.Errorf("启动更新脚本失败：%v", err)
	}
	return nil
}

// currentBundlePath 由可执行文件路径反推 .app 包路径
// （<bundle>.app/Contents/MacOS/<exe> → 上溯三级）
func currentBundlePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("无法定位程序路径：%v", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	bundle := filepath.Dir(filepath.Dir(filepath.Dir(exe)))
	if !strings.HasSuffix(bundle, ".app") {
		return "", fmt.Errorf("当前程序不在 .app 包内（%s），无法自动更新", exe)
	}
	return bundle, nil
}

// findAppBundle 在解包目录里找出 .app 包
func findAppBundle(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), ".app") {
			return filepath.Join(dir, entry.Name()), nil
		}
	}
	return "", errors.New("安装包里没有找到 .app")
}

// cleanupPlatformArtifacts 清理上次更新留下的临时目录。
// 只清理 1 小时前的：刚启动时更新脚本可能仍在执行，删早了会破坏它自己所在的目录。
func cleanupPlatformArtifacts() {
	bundlePath, err := currentBundlePath()
	if err != nil {
		return
	}
	parentDir := filepath.Dir(bundlePath)
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), ".weditor-update-") {
			continue
		}
		info, err := entry.Info()
		if err != nil || time.Since(info.ModTime()) < time.Hour {
			continue
		}
		os.RemoveAll(filepath.Join(parentDir, entry.Name()))
	}
}
