---
status: accepted
---

# 保持 OpenSSH 为运行时权威

Hosta 只发现和索引 Native SSH Config，不重新实现 OpenSSH 的配置求值或 SSH 协议。最终配置由 `ssh -G <alias>` 解析，连接由 `ssh <alias>` 执行，以完整保留 Match、ProxyJump、IdentityAgent 等语义，并保证移除 Hosta 后原生工作流仍然有效。
