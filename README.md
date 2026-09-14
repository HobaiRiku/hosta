# Hosta

> A fast, interactive SSH host launcher built on top of OpenSSH.

Hosta 从现有 OpenSSH 配置发现主机，快速搜索并交由系统 `ssh` 连接。长期目标是在不接管私钥、不破坏原生 SSH 兼容性的前提下，提供可移植的主机资料与自带云目录的跨设备同步。

```text
Discover -> Search -> Select -> Connect
```

项目当前处于 V0.1 开发阶段，尚未发布稳定版本。命令模式、交互式 Launcher、Shell Completion 和跨平台构建链路已经完成。

## 开发预览

需要 Go 1.25 或更高版本：

```bash
git clone git@github.com:HobaiRiku/hosta.git
cd hosta
make build

./hosta
```

也可以直接从源码运行：

```bash
go run ./cmd/hosta
go run ./cmd/hosta list
go run ./cmd/hosta show home
go run ./cmd/hosta doctor
go run ./cmd/hosta connect home
```

生成 Shell Completion：

```bash
hosta completion bash
hosta completion zsh
hosta completion fish
hosta completion powershell
```

`hosta connect <TAB>` 与 `hosta show <TAB>` 会从本地 SSH Config 动态补全 Host，不执行网络请求或 `ssh -G`。

首个 Git Tag 发布后，可从 GitHub Releases 下载 macOS、Linux 和 Windows 的 amd64/arm64 归档及 `checksums.txt`。

读取其他用户配置入口：

```bash
go run ./cmd/hosta --config /path/to/ssh_config list
```

Hosta 元数据使用不会影响 OpenSSH 的结构化注释：

```sshconfig
Host home
    # @hosta.display-name Home Server
    # @hosta.group personal
    # @hosta.tags home server
    # @hosta.description Primary home host
    HostName home.example.com
    User root
```

## 设计文档

- [产品定位与范围](docs/design/01-product-scope.md)
- [系统架构](docs/design/02-architecture.md)
- [数据、SSH Config 与同步协议](docs/design/03-data-and-sync.md)
- [CLI 与交互设计](docs/design/04-cli-and-tui.md)
- [实施计划与验收标准](docs/design/05-delivery-plan.md)
- [设计决策索引](docs/design/README.md)

## 已冻结的核心原则

- OpenSSH First：不实现 SSH 协议，连接始终交给系统 `ssh`。
- SSH Config 是互操作层，不是 Hosta 的同步数据库。
- CLI / TTY First：单一 Go 可执行文件，不依赖 GUI、浏览器或守护进程。
- Local First：没有账号、云服务或 Hosta 配置也能使用核心 Launcher。
- No Secret Sync：私钥、密码、agent socket 和 `known_hosts` 不进入同步域。
- Sync Must Not Block SSH：同步冲突不得阻断本机连接。

## License

[MIT](LICENSE)
