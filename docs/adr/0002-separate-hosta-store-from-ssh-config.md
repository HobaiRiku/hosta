---
status: accepted
---

# 将 Hosta Store 与 Native SSH Config 分离

Native SSH Config 是用户和 OpenSSH 共同使用的互操作层，不承担 Hosta 的稳定实体 ID、版本、同步冲突和 tombstone。Hosta Store 单独拥有这些结构化资料，未来通过显式、可预览的 import/export 与受管 `generated.conf` 互操作，从而避免静默接管用户配置。
