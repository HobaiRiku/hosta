# 数据、SSH Config 与同步协议

## 数据边界

- **Native SSH Config**：用户已有配置和 Include 链；除旧版裸元数据兼容保护外，V0.1 只读。
- **Portable Hosta Data**：V0.2 起由 Hosta 管理，可同步 Host 资料、Group、Description、逻辑 Identity 和版本。
- **Machine Data**：SSH binary、真实私钥路径、终端偏好、设备 ID、缓存、Recent/Frequency、Workspace 本地路径，只保存在本机。

```text
Native SSH Config <-> import/export boundary <-> Portable Store -> Sync Workspace
Machine Data ------------------------------------ local only
```

SSH Config 是互操作层，不等同于 Hosta Store 或同步文件。

## 元数据约定

Hosta 不占用自定义 OpenSSH directive。OpenSSH 会持续增加关键字，例如 `Tag` 已有原生语义；即使使用 `IgnoreUnknown`，自定义参数形态也可能与 OpenSSH parser 冲突。

V0.1 将元数据冻结为 Host block 内的结构化注释：

```sshconfig
Host home
    # @hosta.display-name Home Server
    # @hosta.group personal
    # @hosta.description Primary home host
    HostName home.example.com
    User root
```

OpenSSH 完全忽略这些行，Hosta 按当前 Host block 读取。普通注释不具备语义，`@hosta.*` 之外的名称在 V0.1 也不会进入 Host Index。

为兼容旧版配置，Hosta 仍读取 Host block 内的 `DisplayName`、`Group`、`Description` 裸指令。`Tags` 已移除，不再读取或输出；若旧配置仍包含该裸指令，Hosta 仍会将它加入 `IgnoreUnknown DisplayName,Group,Tags,Description`，保证 OpenSSH 不会因升级而拒绝连接。新格式注释不会触发写入。

## Import / Export

- 显式 alias 可导入；wildcard Host 与 Match 保留为 native-only rule。
- `ssh -G` 可用于预览，不把继承后的所有值无来源地展开写回。
- 私钥仅记录逻辑引用或映射提示，不复制文件。

```text
hosta import --preview / --apply
hosta export --preview / --apply
```

Hosta 永不把整个 `~/.ssh/config` 当生成物。未来只管理 `~/.ssh/hosta/generated.conf`，安装入口 Include 需要明确确认且幂等。

## Portable Host 草案

这是 V0.2 原型输入，不是稳定协议：

```json
{
  "formatVersion": 1,
  "id": "01K5...",
  "revision": "01K5...",
  "parentRevision": "01K4...",
  "alias": "prod-api",
  "displayName": "Production API",
  "ssh": {
    "hostName": "api.example.com",
    "user": "ubuntu",
    "port": 22,
    "identity": "work-default"
  },
  "group": "work",
  "description": "Primary API host",
  "updatedAt": "2026-09-14T10:20:13Z",
  "updatedBy": "device-ulid",
  "contentHash": "sha256:..."
}
```

文件名使用稳定 ULID，不用可变 alias。Revision 使用全局唯一 ID，不能只用离线设备可能撞车的整数。逻辑 identity 在本机映射到真实路径。

## Sync Workspace

V0.3 首个 Backend 是普通目录；iCloud Drive、OneDrive、Google Drive、Dropbox 或 Syncthing 负责跨设备搬运，Hosta 不直接集成 OAuth/API。

```text
Hosta/
  workspace.json
  hosts/<host-id>.json
  groups/<group-id>.json
  revisions/<entity-id>/<revision-id>.json
  tombstones/<entity-id>.json
  conflicts/<entity-id>.json
```

Workspace ID 只标识工作区，不是密钥或认证凭据。一个实体一个文件，canonical serialization 后计算 hash，使用原子替换并保留共同祖先 revision。本地可加进程锁，但不能假设跨设备文件锁有效。

## 云目录探测

`hosta sync detect` 使用平台已知位置、可靠环境变量和明确目录特征。同一 provider 可返回多个账号位置。探测只提供带证据的候选，`sync init` 必须由用户选择或显式传入路径，不能把普通 Documents 目录猜成云目录。

## 合并与冲突

使用 `base/local/remote` 三方合并：

- 不同实体、不同标量字段、双方相同值：自动合并。
- 同一标量字段改成不同值：冲突。
- 集合字段基于 base 计算 add/remove；不矛盾时自动合并。
- 一边删除、一边修改：删除冲突。
- Machine Data 不参与同步。

禁止用 latest/local/remote wins 静默覆盖。删除使用带 parent revision、设备和时间的 tombstone，初步保留 90 天；安全 GC 算法在 V0.3 冻结。

状态为 `synced`、`pending-upload`、`pending-remote`、`conflict`。冲突不阻断本机搜索和连接，用户稍后用 `hosta sync conflicts` 或 `hosta sync resolve` 处理；界面必须展示 base/local/remote，删除冲突不能让 Host 静默复活。

## 安全边界

永不同步私钥、密码、known_hosts、SSH agent socket/token、云盘 OAuth token 和机器绝对私钥路径。HostName 等资料本身也可能敏感，初始化同步时必须明确提示用户数据将以明文进入其选择的第三方目录。

Hosta 的同步层不提供客户端加密。需要加密静态数据的用户应选择自带端到端加密或加密文件系统能力的 Directory Backend；Hosta 不把“传输由云盘负责”误写成“数据已加密”。
