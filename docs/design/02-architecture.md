# 系统架构

## 总体结构

```text
                      Hosta
                        |
          +-------------+-------------+
          |                           |
    SSH Config Source            Hosta Store
    (V0.1 read-only)              (V0.2+)
          |                           |
    Parser + Include Graph       Portable/Machine Data
          |                           |
          +-------------+-------------+
                        |
                    Host Index
                        |
              Search / CLI / TUI
                        |
                 Selected Alias
                  /           \
           ssh -G alias     ssh alias
              show          connect

Hosta Store -> Sync Engine -> Directory Backend -> user's cloud client (V0.3+)
```

UI 依赖应用服务，应用服务依赖领域模型和端口，OpenSSH、文件系统与平台进程是外围适配器。同步不得成为发现、搜索或连接的前置依赖。

## 建议目录

```text
cmd/hosta/main.go
internal/
  app/              application orchestration
  cli/              commands and output formatting
  launcher/         TUI model/update/view
  host/             model, index and search
  sshconfig/        parser, include graph and diagnostics
  openssh/          binary lookup, ssh -G and connection
  platform/         process replacement and OS paths
  store/            portable and machine stores (V0.2+)
  sync/             merge engine and backend ports (V0.3+)
  cloudpath/        cloud-folder detection (V0.3+)
testdata/ssh/
docs/design/
```

目录随里程碑落地，不提前创建空 package。

## SSH Config 子系统

职责：

- 从平台默认路径或 `--config` 入口扫描。
- 保留 directive 顺序、作用域、源文件和行号。
- 展开 `~`、相对路径、多个 Include 参数和 glob。
- 按 OpenSSH 可观察顺序递归遍历，并检测循环。
- 发现显式 alias、Hosta 元数据，输出 Include Graph 和诊断。

它不完整求值 OpenSSH 继承/匹配语义，也不做“后值覆盖前值”的错误 map merge。最终结果交给 `ssh -G`。

按照 OpenSSH 语义，用户配置中的相对 Include 以 `~/.ssh` 为基准，glob 按字典序展开。`${ENV}`、`~`、`%d`、`%u`、`%i`、`%l`、`%L` 和 `%%` 可静态展开；依赖目标 Host 的 token 无法用于全量发现，parser 必须产生诊断而非猜测。

```go
type Node struct {
    Source     SourceLocation
    Directive  string
    Args       []string
    Scope      Scope
}

type SourceLocation struct {
    File string
    Line int
}
```

Host 发现规则：

- 只收录不含通配符且未被 `!` 否定的显式 pattern。
- `Host *`、wildcard 和 `Match` 保留在来源模型，但不进入 Launcher。
- 同一 alias 多次声明时保留全部来源；索引层只产生一个候选并报告重复。

## OpenSSH 适配器

- `Locate()` 查找 `ssh` / `ssh.exe`。
- `Resolve(alias)` 安全执行 `ssh -G` 并解析所需字段。
- `Connect(alias, args)` 将当前终端交给 OpenSSH。

Unix 优先进程替换以继承 TTY、signals、窗口尺寸和退出码；Windows 通过子进程继承 stdio 并透传退出码。所有调用使用参数数组，不经 shell 拼接；alias 必须按命令合约防止选项注入。

## Host Index 与搜索

```go
type Host struct {
    ID          string
    Alias       string
    DisplayName string
    Group       string
    Description string
    Preview     Preview
    Sources     []SourceLocation
    Origin      Origin
}
```

Preview 仅来自轻量解析，允许不完整；`show` 才调用 `ssh -G`。默认搜索权重：Alias 100、DisplayName 90、Group 60、HostName Preview 50、Description 30。算法必须正确处理 Unicode，不能按 UTF-8 byte 评分。

## 错误与非功能约束

建议退出码：输入错误 2、配置错误 3、OpenSSH 不可用 4、resolve 失败 5、Ctrl+C 在 Unix 为 130；连接尽量返回 OpenSSH 的退出码。实现首个命令时以测试冻结。

- 启动路径不访问网络。
- 日志不得输出秘密或完整敏感环境变量。
- 未来配置写入必须同目录临时文件 + flush/close + atomic rename，并提供恢复备份。
- parser、merge engine 和平台边界必须有独立测试与 fixture。
