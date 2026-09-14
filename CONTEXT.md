# Hosta

Hosta 在不替代 OpenSSH 的前提下，为 SSH Host 提供发现、检索、资料管理与可选的跨设备同步。

## Language

**SSH Host**:
用户可以交给 OpenSSH 解析和连接的命名目标，以 alias 标识。
_Avoid_: Server, machine

**Native SSH Config**:
由用户和其他工具拥有、供 OpenSSH 读取的配置入口及其 Include 链；Hosta 默认只读。
_Avoid_: Hosta config, database

**Hosta Store**:
Hosta 拥有的结构化持久化数据，保存可管理的 SSH Host 资料，但不保存秘密。
_Avoid_: SSH Config, sync folder

**Portable Data**:
允许在设备间复制的 Hosta Store 数据，包括 SSH Host 资料和逻辑 Identity 引用。
_Avoid_: cloud config

**Machine Data**:
仅属于一台设备的数据，包括真实私钥路径、设备偏好、缓存和本地统计。
_Avoid_: local override

**Identity Reference**:
Portable Data 中引用某类 SSH 身份的逻辑名称，在每台设备上分别映射到真实 IdentityFile。
_Avoid_: private key

**Sync Workspace**:
承载 Portable Data 的目录及其稳定 workspace ID，可由用户选择的第三方工具复制到其他设备。
_Avoid_: cloud account, configuration key

**Directory Backend**:
将 Sync Workspace 视为普通文件系统目录的同步后端；云盘客户端或其他工具负责目录的远端传输。
_Avoid_: cloud API

**Revision**:
一个实体内容版本的全局唯一标识，并通过 parent revision 建立可用于三方合并的版本关系。
_Avoid_: integer version

**Tombstone**:
代表实体已被删除且仍需传播给其他设备的版本化记录。
_Avoid_: deleted file

**Conflict**:
基于共同祖先无法无损自动合并的两个变更；它属于同步状态，不会阻断 SSH Host 的本地使用。
_Avoid_: latest wins
