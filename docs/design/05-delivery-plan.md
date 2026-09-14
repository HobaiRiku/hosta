# 实施计划与验收标准

## 策略

先建立可运行的纵向链路，再扩展管理与同步。每阶段都可测试、可演示，不创建尚未使用的空 package。

## 当前进度

| 里程碑 | 状态 | 说明 |
| --- | --- | --- |
| M0 Bootstrap | 已完成 | Go module、技术栈、命令/TUI 骨架、测试、三平台 CI |
| M1 SSH Config Parser | 已完成 | AST、Include Graph、诊断与 Host discovery |
| M2 Host Index/Search | 已完成 | 统一模型、字段权重、确定性排序与基准测试 |
| M3 CLI/OpenSSH | 已完成 | list/show/doctor/connect/config、resolve 与平台进程边界 |
| M4 Interactive Launcher | 已完成 | 实时检索、导航、详情、TTY/小终端降级与选择后连接 |
| M5 Completion/Release | 已完成 | 四种 Shell 补全、跨平台归档、校验和与 Tag 发布工作流 |

V0.1 代码与本地发布快照已经完成。正式对外发布仍需确认首个版本号、推送提交，并创建对应 Git Tag；这些属于发布动作，不由实现阶段自动执行。

## V0.1

### M0：Bootstrap

- 初始化 Go module、命令入口、格式化/lint/test。
- 冻结 Go 版本、平台矩阵；验证 Cobra/Bubble Tea 与进程模型。
- 建立三平台 build/test CI 与启动 benchmark 骨架。

验收：三平台可构建，`hosta version` 可运行，检查有统一入口。

冻结技术栈：Go 1.25、Cobra 1.x、Bubble Tea/Bubbles/Lip Gloss v2、`sahilm/fuzzy`。第三方依赖使用明确版本，由 Go module checksum 校验。

### M1：SSH Config Parser

- AST/source location、Host/Match scope。
- `~`、相对路径、glob、多参数、递归 Include。
- 循环/缺失/权限/重复 alias 诊断和元数据读取。
- macOS/Linux/Windows fixture。

验收：fixture 产生确定 Host 集与 Include Graph，诊断有文件/行号，不修改真实配置。

### M2：Host Index/Search

- 统一读模型、去重、Unicode fuzzy、权重、确定性排序和 benchmark。

验收：中英文、前缀、空查询、重复 alias 有测试，目标规模搜索无可感知停顿。

### M3：CLI/OpenSSH

- `list/ls`、`show`、`doctor`、`connect/c`。
- OpenSSH 查找、`ssh -G`、错误/stdio/signal/exit code 映射。
- JSON 合约测试。

验收：`show` 与同环境 `ssh -G` 一致；参数不经 shell；平台边界可审计且有测试。

### M4：Interactive Launcher

- 搜索、导航、详情、退出、连接。
- 小终端、无颜色、无 TTY、空 Host 降级。
- 启动/渲染 benchmark。

验收：键盘契约有 Model/Update 测试；启动不访问网络、不逐 Host resolve。

### M5：Completion/Release

- 四种 shell completion、跨平台产物/校验和、安装说明、changelog。
- 脱敏真实配置冒烟测试。

验收：全平台 CI 通过，README 流程可复现，安全与范围清单通过。

## V0.2：Store

实现前用原型冻结 Portable/Machine schema、ID/revision、import/export round-trip、generated.conf 所有权与恢复。只有 preview/diff、备份、幂等和恢复测试齐全后才允许 `--apply` 写 SSH 目录。

## V0.3：同步

1. 先实现两个本地目录间的纯同步算法。
2. revision graph、三方字段合并和 tombstone。
3. 冲突持久化与非阻塞交互。
4. 测试异常写、重复事件、部分下载和冲突副本。
5. 云目录探测与确认。
6. 真实 iCloud/OneDrive/Google Drive 人工矩阵。

不得先写云厂商 API 来绕过合并正确性。

## 测试策略

- Parser：table tests、fixtures、fuzz。
- Search：Unicode、稳定排序、property/benchmark。
- CLI：golden output、exit code、non-TTY。
- OpenSSH：fake executable contract + 少量本机 integration。
- TUI：Model/Update 为主，PTY smoke 为辅。
- Sync：状态矩阵、崩溃恢复、幂等、并发、随机序列。
- Security：argument injection、path traversal、symlink、权限、日志泄密。

## Definition of Done

- 行为符合设计；偏离先更新决策。
- 新逻辑有与风险相称的测试。
- test、format、static check、目标平台 build 通过。
- 不修改或泄露用户真实 SSH 配置与私钥。
- README/帮助同步更新，不夹带后续里程碑功能。

## 开工后的第一批任务

1. 初始化 Go module 与 `cmd/hosta`，完成 M0。
2. 建立 SSH Config fixture，实测 OpenSSH Include 顺序与边界。
3. 设计最小 AST，用测试驱动 Host discovery。
4. Parser 可用后实现 Index/Search，再接 CLI/TUI。

本轮仅落地设计文档；业务代码从用户明确确认后开始。
