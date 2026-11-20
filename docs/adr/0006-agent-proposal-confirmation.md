# Require confirmation before Agent canvas writes

- Status: accepted
- Date: 2026-11-19

Agent 只能读取项目上下文并返回说明和结构化 Operation 提议；用户在画布中预览并确认后，客户端才提交 Operation。Agent Run 的完成不等于画布变更完成。这个选择把模型的不确定性隔离在提议阶段，避免一次错误输出直接破坏用户项目，也使“模型回答”和“持久化写入”可以独立重试和审计。

## Considered Options

- **Agent 自动执行**：交互更短，但误操作不可逆，且模型输出无法在提交前得到用户确认。
- **只返回文本，不允许结构化变更**：安全但失去 Agent 操作画布的核心价值，用户必须手工重复执行。

## Consequences

- Web 必须实现 Operation 预览、确认、放弃和冲突提示状态。
- Backend 必须区分 `operation.proposed` 与实际提交成功，不能在模型调用结束时隐式写库。
- 后续如果增加自动执行模式，必须作为显式权限和独立产品决策，不能复用默认流程暗中开启。
