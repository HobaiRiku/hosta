---
status: accepted
---

# 使用不加密的 Directory Backend 同步

Hosta 的首个同步后端只读写普通目录，由 iCloud Drive、OneDrive、Google Drive、Dropbox 或其他用户工具负责远端传输；Hosta 不集成厂商 API，也不提供加密层。Portable Data 将以明文保存，因此秘密永不进入同步域，需要静态加密时由用户选择具备相应能力的目录后端。
