# WEditor 自动更新 — 设计

日期：2026-10-10
状态：已实施

## 背景

WEditor 此前没有任何版本检测机制：`app.go` 里 `APP_VERSION` 是写死的 `1.0.0`，CI 只把版本写进 `frontend/package.json`，macOS 包信息与 Windows 版本资源也都是固定的 `1.0.0`。也就是说无论打什么 tag，安装后的 App 都自报 1.0.0，**没有任何判断新旧的基础**。

用户诉求：GitHub 打包成功后，对应系统的客户端能自动检测到新版本，然后下载、更新、重启。

## 目标

1. 构建时把真实版本号注入产物（Go 变量 + macOS plist + Windows 版本资源）
2. App 能判断"有没有新版本"，并且判断依据是发布方权威声明的、而不是猜出来的
3. 用户点一次即可完成 下载 → 校验 → 替换 → 重启
4. 失败时绝不留下半残状态（下坏不装、装失败回滚、无权限则降级为手动）

非目标：Linux 自动安装（仅检测+下载+引导手动安装）、代码签名与公证、增量更新（每次下载完整包）。

## 关键决策与理由

| 决策 | 理由 |
|---|---|
| CI 产出 `latest.json` 清单，App 读清单（而非直接查 GitHub API） | 未认证 API 每小时 60 次（NAT 下多用户共享会打爆）；清单能声明资产名与 sha256，不必猜文件名，也可校验完整性 |
| 稳定通道读 `/releases/latest/download/latest.json` | 静态重定向，不消耗 API 配额；且自动跳过 prerelease —— CI 在 main 分支发的 dev 构建正是 prerelease，语义刚好吻合 |
| 开发通道查 Releases API 取 `published_at` 最新者 | dev 版本形如 `v1.0.0-dev-<sha>`，sha 之间**没有大小关系**，semver 排不出先后，只能用发布时间定序 |
| 稳定通道用 semver 比较，其中两个 dev 版本视为相等 | 同上；相等后由通道策略决定，不会误触发 |
| macOS 用独立脚本在 App 退出后替换整个 `.app` | 把新包搬到位后本进程仍持有旧 inode；只有退出后再操作才干净 |
| macOS 用 `ditto` 解包，临时目录与目标同级 | `ditto` 能原样保留权限位与符号链接（`archive/zip` 会丢）；同级目录保证最后的 `mv` 是同分区 rename，原子且不丢元数据 |
| Windows 先重命名运行中的 exe，再放入新版本 | Windows 允许重命名运行中的 exe，但不允许覆盖/删除它 |
| 不做旧版本备份回滚 | 新版本启动失败时用户根本无法运行 App 去触发回滚，备份救不了；安装包仍在本地，`打开下载目录` + Release 页足够恢复 |

## 版本流

**构建期注入**（`build.yml`）：

```
VERSION    = tag 名（如 v1.4.2）或 v1.0.0-dev-<sha>
NUMERIC    = VERSION 去掉 -后缀 的数值部分（macOS plist / Windows 版本资源要求 x.y.z）
COMMIT, BUILD_TIME = git short sha / UTC RFC3339
LDFLAGS    = -X main.APP_VERSION=$VERSION -X main.APP_COMMIT=$COMMIT -X main.APP_BUILD_TIME=$BUILD_TIME
```

- `app.go`：`APP_VERSION` 由 `const` 改为 `var`，保留默认值 `1.0.0` 供本地 `wails dev` 使用；新增 `APP_COMMIT`、`APP_BUILD_TIME`
- `wails.json → info.productVersion` 与 `build/darwin/Info.plist` 的 `CFBundleVersion`/`CFBundleShortVersionString` 同步为 `NUMERIC`
  - **注意**：macOS 的 `build/darwin/Info.plist` 是原样拷贝的（不含 wails 模板变量），必须直接改写文件；Windows 的 `build/windows/info.json` 才会展开 `{{.Info.ProductVersion}}`
- CI 有一步 `Verify build metadata`，`LDFLAGS` 为空直接让 job 失败——否则版本静默停在 1.0.0，自动更新会永远失效

**注入形式必须是 `-X main.APP_VERSION`**：对真实二进制有效；`-X mdeditor.APP_VERSION`（模块名）只对 `go test` 的测试二进制有效。

**比较规则**：

- 稳定通道：`newerThan(latest, current)`，即 semver 比较（`1.0.0` > `1.0.0-dev-x`）
- 开发通道：`publishedAt > APP_BUILD_TIME`；`APP_BUILD_TIME` 为空的本地构建不做判断
- 两条通道都遵守 `skippedVersion`（跳过此版本）
- 任一版本号无法解析时一律判定为"不更新"——宁可漏报，不可误报触发一次破坏性替换

## 清单契约（CI ↔ App）

CI 的 `Generate update manifest` 步骤产出并随 Release 上传：

```json
{
  "version": "v1.4.2",
  "channel": "stable",
  "publishedAt": "2026-10-10T07:40:18Z",
  "notes": "Release v1.4.2",
  "assets": {
    "darwin-universal": { "name": "...", "url": "...", "size": 0, "sha256": "..." },
    "windows-amd64":    { "name": "...", "url": "...", "size": 0, "sha256": "..." },
    "linux-amd64":      { "name": "...", "url": "...", "size": 0, "sha256": "..." }
  }
}
```

App 侧键名匹配顺序：`<GOOS>-<GOARCH>` → `darwin-universal`（macOS 回退）。该契约由 `TestManifestContract` 固定，两侧单方面改动会立刻测试失败。

## 运行前提（不满足则功能整体失效）

**Release 资产必须能被匿名访问。** App 不携带任何 GitHub 凭据，取清单、下产物用的都是未登录身份。

- 仓库私有 → `releases/latest/download/latest.json` 返回 404、Releases API 返回 403，
  自动更新**整体失效**（App 只会显示"检查更新失败"）。这不是降级，是完全不可用；
  且失败是静默的（自动检查不弹错），很容易误以为功能正常。
- 本仓库最初就是私有，发布 v1.1.0 前已改为公开。
- 若将来必须私有：需要把清单与产物发布到任何可匿名访问的位置（公开仓库 / 对象存储 /
  静态站点），并同步改三处——`stableManifestURL`、CI 生成清单时的资产 URL、
  开发通道的 API 查询地址。

**开发通道受 GitHub API 未认证限流**（60 次/小时/IP）。稳定通道走静态地址不受影响。
进程内 ETag 缓存只能减少同一会话内的重复检查，每次重启仍消耗一次。

## 组件

新增文件：

- `version.go` — semver 解析/比较（纯函数）
- `updater.go` — 清单获取、通道策略、资产匹配、下载与 sha256 校验、设置持久化、更新流程编排
- `updater_apply_darwin.go` / `updater_apply_windows.go` / `updater_apply_other.go` — 各平台安装与残留清理
- `frontend/src/composables/useUpdater.js` — **模块级单例**状态机（idle/checking/available/downloading/ready/installing/error）
- `frontend/src/components/UpdateDialog.vue` — 更新弹窗
- `frontend/src/components/settings/UpdateSection.vue` — 设置面板「自动更新」页

绑定方法（Go → 前端）：`CheckForUpdate(manual)`、`DownloadUpdate()`、`ApplyUpdateAndRestart()`、`GetUpdateSettings()`、`SetUpdateSettings()`、`SkipVersion()`、`RevealUpdateFile()`、`OpenUpdateDownloadPage()`、`GetBuildInfo()`。
事件：`update-progress`。

设置持久化在 `<AppData>/update.json`：`{channel, autoCheck, skippedVersion, lastCheckAt}`。默认 `stable` + 自动检查——普通用户不该被 dev 构建打扰。

## 平台安装

**macOS**：定位 bundle（`os.Executable()` 上溯三级）→ 开发构建守卫 → 目标目录可写预检 → 同级临时目录 + `ditto -x -k` 解包 → 写 `apply.sh` 并以 `Setsid` 分离启动 → App 延迟 800ms 退出 → 脚本等父进程退出后：旧包移入临时目录、新包搬到位、清隔离属性、`open -n` 重启；搬移失败则把旧包还原，至少让用户还有可用版本。

**Windows**：开发构建守卫 → 目录可写预检 → 旧的 `WEditor.exe.old` 清理 → 重命名当前 exe 让出文件名 → 复制新版本到位（失败则回滚重命名）→ 拉起新进程 → 自身退出；`.old` 在下次启动时清理。

**其他平台**：返回"暂不支持自动更新"，前端降级为「打开下载目录 / 打开发布页面」。

**开发构建守卫**：可执行文件向上 8 级内存在 `.git` 即判定为开发构建，拒绝自动替换，避免把 `build/bin` 下的开发产物覆盖掉。

## 错误处理

- 自动检查失败静默（只写日志），手动检查才向用户报错
- 所有元数据请求 15s 超时；安装包下载用独立的长超时（30 分钟）
- 下载流式落盘并同时计算 sha256，不匹配即删除 `.part` 并报错，**绝不安装**
- 安装前再次校验 sha256（下载与安装之间可能被改动）
- 无写权限时不硬来，降级为手动安装
- 启动时清理：`.part` 立即删除；其余残留（含 macOS 备份目录）超过 24 小时清理；macOS 只清理 1 小时前的临时目录（刚启动时更新脚本可能仍在运行）

## 与初版设计的两处偏差

1. **触发启动检查的位置**：改由前端在注册完事件监听后延迟 5s 调用 `CheckForUpdate(false)`，而不是后端 `OnDomReady` 后推事件。这样彻底消除"事件早于监听"的竞态，且行为一致。
2. **未保存内容警告**：原计划用脏标记提示"更新会丢内容"。实施时发现项目没有脏标记，但**已有 localStorage 草稿自动保存**（重启后可恢复），因此改为如实提示"未保存的编辑已自动存为草稿"，不为一个提示去新增脏跟踪。

## 测试与验证

- `version_test.go`：semver 解析、比较、`newerThan` 的保守性（解析失败不更新）
- `updater_test.go`：开发通道选取、资产匹配优先级、时间比较、开发构建识别、ETag 304 缓存、下载校验和、设置往返、跳过版本、**清单契约**、以及 `TestUpdateCheckAndDownloadEndToEnd`（本地假清单跑通 检查→下载→校验→复用→已最新→已跳过）与 `TestDownloadUpdateRejectsTamperedPackage`（校验失败不留残留）
- 已实测确认：`-X main.APP_VERSION` 对真实二进制生效（临时探针读出 `v9.9.9-AAA|aaa1111|2026-10-10T00:00:00Z`）
- 未自动化：平台替换与重启（无法单元测试），需按下列步骤人工验收

## 人工验收步骤

1. 关闭「稳定版通道」以外的干扰，确认关于页/设置页显示真实版本号
2. 打一个 tag（如 `v1.1.0`）触发 CI，确认 Release 里出现 `latest.json` 且 sha256 与资产一致
3. 在装有旧版本的机器上启动 App，5 秒内应弹出更新提示
4. 点「下载并安装」，观察进度→重启，重启后设置页版本号应为新版本
5. 降级路径：把 App 放到只读位置（如 DMG 内）触发更新，应给出明确提示而非静默失败
