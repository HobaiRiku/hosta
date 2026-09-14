# 产品定位与范围

## 一句话定义

Hosta 是建立在现有 OpenSSH 配置之上的快速、交互式、跨平台 SSH Host Launcher。

```text
读取配置 -> 发现 Host -> 搜索/选择 -> exec ssh <alias>
```

它不替代终端，不实现 SSH，也不要求用户放弃 `ssh`、`scp`、`rsync` 等原生命令。

## 目标

- 零配置读取已有 SSH Host。
- 在终端中以极短路径完成搜索和连接。
- 显示 Host 的来源与 OpenSSH 最终生效配置。
- 逐步提供 DisplayName、Group、Tags、Description 等资料。
- 未来借助已有云盘目录同步可移植资料，不建设 Hosta 云端账号体系。

## 产品原则

- **OpenSSH First**：连接执行 `ssh <alias>`；最终有效配置通过 `ssh -G <alias>` 获取；复杂语义由 OpenSSH 负责。
- **Native Compatibility**：删除 Hosta 后，原有 SSH 工作流仍然有效；不持续重写入口配置。
- **Local / TTY First**：核心能力不需要 GUI、Daemon、账号或联网。
- **Fast Launcher**：启动时禁止逐 Host 做 DNS、Ping、TCP、`ssh -G` 或 SSH Test。冷启动目标 50 ms 内，可接受上限 100 ms，最终以基准校准。
- **Safe by Boundary**：不保存密码、不同步私钥、不代理 SSH 会话。未来配置写入必须显式、可预览、原子化。

## 版本范围

### V0.1：SSH Launcher

- 发现默认 SSH Config 及递归 Include。
- 解析显式 Host alias 和 Hosta 元数据。
- Host Index、Unicode 模糊搜索和交互式 Launcher。
- `connect/c`、`list/ls`、`show`、`doctor`、`completion`、`version`。
- 使用系统 OpenSSH；覆盖 macOS、Linux、Windows。

V0.1 对 SSH Config 只读，不创建 Hosta Store，不做同步。

### V0.2：可管理资料

- Hosta Store 与稳定 Host ID。
- 显式 import、preview、apply/export。
- Portable Data 与 Machine Data 分层。
- Group、Tags、DisplayName、Description、Recent、Favorite。

### V0.3：目录同步

- Directory Backend 与常用云盘目录探测。
- Workspace 初始化/加入、增量扫描、三方合并、tombstone 和冲突处理。
- 同步状态可见，但不得阻断 SSH。

### V1.0 后再评估

Git/WebDAV/云厂商 Backend、Host 编辑器、Quick Actions、远程命令、状态缓存、Plugin。

## 明确不做

- SSH 协议或 Go SSH Client、Terminal Emulator、SFTP、GUI。
- 密码、私钥、known_hosts、SSH agent socket 的存储或同步。
- 默认网络检查和服务器监控。
- V0.1 的 Host 创建、删除、编辑与同步。
- 静默接管或整体覆盖 `~/.ssh/config`。

## 成功标准

- 安装后运行 `hosta` 即可搜索现有 Host。
- 连接行为与直接运行 `ssh <alias>` 一致。
- Include 错误可定位到文件和行号。
- 一般规模配置达到启动性能预算。
- 三个平台的核心命令语义一致。
