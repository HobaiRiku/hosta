# CLAUDE.md

本文件为在本仓库工作的 AI 编码助手（Claude Code、Codex 等）提供项目约定。`AGENTS.md` 直接引用本文件，只维护这一份。

## 项目概述

Hosta 是基于 OpenSSH 的交互式 SSH Host 启动器（Go CLI + TUI）：从用户现有的 SSH config（含 `Include` 链）发现 Host，提供模糊搜索、Shell 补全，并把所有连接委托给系统 `ssh`。

- 模块路径：`github.com/HobaiRiku/hosta`，Go 1.25
- 技术栈：Cobra、Bubble Tea / Bubbles / Lip Gloss v2（`charm.land/*`）、`sahilm/fuzzy`、`atotto/clipboard`
- 平台：Linux、macOS、Windows（amd64 / arm64），`CGO_ENABLED=0`
- 当前阶段：V0.1（M0–M5）已完成；Hosta Store（V0.2）与目录同步（V0.3）仍处于设计阶段

## 常用命令

```bash
make build        # 构建到 build/bin/hosta，并注入版本信息
make run          # go run ./cmd/hosta
make test         # go test ./...
make vet          # go vet ./...
make fmt          # gofmt -w
make check        # fmt-check + vet + test + build，提交前必须通过
go test ./internal/sshconfig -run TestParseIncludesAndDiscoverHosts   # 运行单个包/单个测试
go run ./cmd/hosta --config testdata/ssh/config list   # 用 fixture 手动验证
```

CI（`.github/workflows/ci.yml`）在三平台执行 `go test`、`go vet`、`go build`，并单独检查 `gofmt -l`。

## 目录结构

```text
cmd/hosta/main.go        入口：cli.Execute()，按 cli.ExitCode(err) 退出
internal/
  cli/                   Cobra 命令、输出格式、错误码（errors.go）
  app/                   应用编排：加载 SSH config 并构建 Snapshot/Index
  sshconfig/             解析器、Include 图、token 展开、诊断、旧元数据兼容
  host/                  Host 模型、索引、加权搜索
  search/                Unicode 模糊匹配
  launcher/              Bubble Tea TUI（model/update/view）
  openssh/               ssh 二进制查找、ssh -G 解析、连接
  platform/              进程替换 / 交互终端的平台差异（_unix / _windows）
  version/               由 ldflags 注入的版本元数据
testdata/ssh/            parser fixture
docs/design/             产品范围、架构、数据同步、CLI/TUI、交付计划
docs/adr/                架构决策记录
CONTEXT.md               领域术语表
```

分层方向：`cli` / `launcher` → `app` → `host` / `sshconfig` / `search`；`openssh`、`platform`、文件系统属于外围适配器。目录随里程碑落地，**不要提前创建空 package**（如 `store/`、`sync/`）。

## 核心架构约束

以下约束来自 `docs/adr/` 与 `docs/design/02-architecture.md`，修改前务必遵守：

1. **OpenSSH 是运行时权威**（ADR 0001）：不重新实现 OpenSSH 配置求值或 SSH 协议。最终配置用 `ssh -G <alias>`，连接用 `ssh <alias>`；移除 Hosta 后原生工作流必须仍然可用。
2. **Native SSH Config 默认只读**（ADR 0002）：Hosta 自有数据将来放在独立的 Hosta Store。目前唯一的写入是 `sshconfig.EnsureLegacyCompatibility` 为旧式元数据插入 `IgnoreUnknown`，且必须走同目录临时文件 + atomic rename（`replace_unix.go` / `replace_windows.go`）。
3. **Parser 不做 map merge**：保留 directive 顺序、作用域、源文件和行号；相对 Include 以 `~/.ssh` 为基准，glob 按字典序展开；依赖目标 Host 的 token 无法静态展开时产生诊断而非猜测。
4. **Host 发现规则**：只收录不含通配符且未被 `!` 否定的显式 pattern；`Host *`、wildcard、`Match` 保留在来源模型中，但不进入 Launcher。重复 alias 保留全部来源，索引层只产生一个候选。
5. **Host 元数据**使用 `# @hosta.*` 结构化注释；旧的 `DisplayName/Group/Description` 自定义 directive 仅做兼容读取。旧的 `Tags` directive 仅保留 OpenSSH 忽略保护，不再进入 Host 数据。
6. **搜索**：权重 Alias 100、DisplayName 90、Group 60、HostName Preview 50、Description 30；排序必须确定；按 rune 而非 UTF-8 byte 评分。
7. **进程调用安全**：外部命令一律使用参数数组，不经 shell 拼接；alias 需防止被当作选项注入。
8. **进程模型**：Unix 优先进程替换以继承 TTY、信号、窗口尺寸和退出码；Windows 使用子进程继承 stdio 并透传退出码。平台差异通过 `_unix.go` / `_windows.go` 构建标签隔离。
9. **启动路径不访问网络**；日志/输出不得包含秘密或敏感环境变量；秘密永不进入未来的同步域（ADR 0003）。

## 错误与退出码

错误通过 `internal/cli/errors.go` 的 `withCode(code, err)` 包装，`ExitCode` 统一映射，已由测试冻结：

| 码 | 含义 |
| --- | --- |
| 2 | 输入 / 参数错误 |
| 3 | 配置错误 |
| 4 | OpenSSH 不可用 |
| 5 | `ssh -G` resolve 失败 |
| 130 | Unix 下 Ctrl+C |
| 其他 | 连接时透传 OpenSSH 的退出码 |

## 编码约定

- 遵循 Go 惯用法：显式返回 `(value, error)`，用 `fmt.Errorf("...: %w", err)` 附带上下文，不使用 panic 做流程控制。
- 条件分支优先 guard clause 提前返回；多分支映射优先使用 map / 表驱动，而非长 `else if` 链。
- 命令通过 `dependencies`（`load`、`newSSH`）注入外部依赖，测试时替换为 fake，不在单元测试中调用真实 `ssh` 或修改真实 `~/.ssh/config`。
- 测试以表驱动为主，parser、索引、搜索、平台边界需有独立测试与 `testdata/` fixture；新增行为同步补测试。
- 保持 `gofmt` 干净；新增第三方依赖需固定明确版本并执行 `go mod tidy`。
- 用户可见的 CLI 输出与 README 使用英文；设计文档、ADR、`CONTEXT.md` 使用中文。领域术语以 `CONTEXT.md` 为准（如使用 "SSH Host" 而非 "Server"）。

## 文档维护

- 用户可见的新增或变更记录到 `CHANGELOG.md` 的 `[Unreleased]`（Keep a Changelog 格式）。
- 命令或快捷键变化时同步更新 `README.md` 与 `docs/design/04-cli-and-tui.md`。
- 做出新的架构取舍时在 `docs/adr/` 新增编号 ADR；里程碑进度更新 `docs/design/05-delivery-plan.md`。
- Markdown 中代码块的开始与结束标记前后保留空行（兼容钉钉文档）。

## 提交与发布

- 提交信息使用 Conventional Commits 前缀 + 中文描述，例如 `feat: 支持复制 SSH 命令到剪贴板`、`fix: ...`、`docs: ...`、`build: ...`。
- 发布由推送 Git Tag 触发 GoReleaser（`.goreleaser.yml`、`.github/workflows/release.yml`），同时发布 Homebrew Tap。打 Tag、推送、发布均属于发布动作，需用户明确要求后才执行。
