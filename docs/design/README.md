# Hosta 设计文档

状态：V0.1 implementation baseline
基线日期：2026-09-14

本目录是 Hosta 开发的当前设计基线。架构为未来演进预留边界，但未列入当前里程碑的能力不能提前进入开发。

## 阅读顺序

1. [产品定位与范围](01-product-scope.md)
2. [系统架构](02-architecture.md)
3. [数据、SSH Config 与同步协议](03-data-and-sync.md)
4. [CLI 与交互设计](04-cli-and-tui.md)
5. [实施计划与验收标准](05-delivery-plan.md)

## 决策状态

| 编号 | 决策 | 状态 |
| --- | --- | --- |
| D-001 | 使用 Go 构建单一跨平台二进制 | 已确定 |
| D-002 | 使用系统 OpenSSH 解析最终配置并建立连接 | 已确定 |
| D-003 | V0.1 以现有 SSH Config 为只读 Host 来源 | 已确定 |
| D-004 | Hosta Store 与 SSH Config 分离 | 已确定 |
| D-005 | 同步首选 Directory Backend，不直接接云厂商 API | 已确定 |
| D-006 | 同步采用逐实体文件、稳定 ID、三方合并和 tombstone | 已确定；具体格式由 V0.3 原型冻结 |
| D-007 | Cobra 1.x + Bubble Tea/Bubbles/Lip Gloss v2 | 已确定 |
| D-008 | V0.1 使用 `sahilm/fuzzy`，在 Host 层叠加字段权重 | 已确定 |
| D-009 | Store 持久化格式使用 JSON | 已确定；schema 由 V0.2 原型冻结 |
| D-010 | 同步层不提供数据加密 | 已确定 |
| D-011 | SSH Config 元数据使用 `# @hosta.*` 结构化注释 | 已确定 |

## 变更规则

- 改变核心原则、数据所有权或里程碑范围时，先更新文档并记录理由。
- 实现细节可随原型调整，但不得自行扩大产品边界。
- 示例格式不是稳定公开协议；只有明确标记 `formatVersion` 的格式才进入兼容承诺。
