# Use Canvas Operation as the only canvas write language

- Status: accepted
- Date: 2026-11-19

画布的所有写入（用户手势、快捷键和 Agent 结果）统一表达为 `Canvas Operation` 批次，并由同一个 domain validator 校验和应用。我们拒绝让前端直接拼接快照或让 Agent 直接调用 repository，因为统一写入语言可以集中处理节点/边约束、级联删除、幂等和审计，同时让前端 reducer、后端事务和 Agent 输出共享同一套语义。

## Considered Options

- **直接修改 CanvasProject 快照**：接口简单，但规则会散落在多个调用方，Agent 和用户操作容易产生不一致。
- **完整事件溯源**：审计能力更强，但一期需要额外的事件存储、重放和投影复杂度，超出当前需求。

## Consequences

- `Canvas Operation` 成为前后端共享的领域术语和测试接口。
- Operation 批次必须具备原子性；任意一项失败时整批不提交。
- 后续新增节点类型或规则需要扩展 validator，而不是绕过它增加特殊写入口。
