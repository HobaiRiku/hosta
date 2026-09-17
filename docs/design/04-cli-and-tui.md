# CLI 与交互设计

## 命令模型

无参数进入 Launcher，而不是显示帮助：

```text
hosta
hosta connect <host> [-- <ssh-args...>]
hosta c <host> [-- <ssh-args...>]
hosta list | hosta ls
hosta show <host>
hosta config
hosta doctor
hosta completion <shell>
hosta version
```

不支持 `hosta <host>`，避免 alias 与未来子命令冲突。V0.2/V0.3 的 import/export、sync、edit、test 等命令暂不注册。

## Launcher

```text
 Hosta                                         16 hosts

 Search  rou|

 > MT2500 Router
   mt2500            personal · home · router
   root@192.168.90.1:322

 1 match
```

| 输入 | 行为 |
| --- | --- |
| 可打印字符 | 写入搜索框 |
| Backspace/Delete | 删除一个用户感知字符 |
| Up / Ctrl+P | 上一个结果 |
| Down / Ctrl+N | 下一个结果 |
| Enter | 连接当前 Host 并离开 Hosta |
| Ctrl+] | 进入 Group 名称输入过滤；输入时实时筛选，Enter 保留筛选并返回 Host 选择 |
| Esc | 有查询时清空；无查询时退出 |
| Ctrl+C | 退出 |
| Tab | 展开/收起详情 |

默认不绑定 `j/k` 导航，因为始终处于搜索输入模型。非交互环境运行裸 `hosta` 时返回清晰错误，并提示 `hosta ls` 或 `hosta c`。

## 搜索与排序

查询 Alias、DisplayName、Group、HostName Preview、Description。空查询按 Group、Alias 稳定排序；非空先按 score，再稳定打破平局。支持 ASCII 大小写不敏感、中文/Unicode、连续与词首命中、Alias 精确前缀加权。V0.1 不加入 Recent/Favorite 权重。

## 核心输出

`list` 默认人读表格，同时提供以 golden test 冻结的 `--json`。

`show` 分开显示 Hosta 元数据、OpenSSH 最终配置与来源；最终配置标明由 OpenSSH resolve。敏感或冗长字段默认不全量打印。

`doctor` 只诊断本机环境和静态配置：OpenSSH、入口、Include、循环/缺失、重复 alias、结构化元数据、可明确判断的 IdentityFile。它不探测远端网络。

## Completion

- `hosta <TAB>` 补子命令，不补 Host。
- `hosta c/show <TAB>` 补 alias，以 DisplayName 为描述。
- 补全过程不执行 `ssh -G` 或网络请求。
- V0.1 覆盖 bash、zsh、fish、PowerShell。

## 连接语义

Enter 或 `hosta c home` 是 Launch and Leave Hosta。Launcher 选择后会先显示 `Connecting to <alias> ...`；OpenSSH 成功建立会话后，通过受控的本地命令清屏，使远端会话从干净终端开始。`--` 后参数透传及 alias/参数选项注入规则必须以安全测试冻结。
