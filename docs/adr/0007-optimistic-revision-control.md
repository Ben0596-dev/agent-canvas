# Use optimistic revision control for project writes

- Status: accepted
- Date: 2026-11-19

每个 Canvas Project 维护单调递增的 `revision`。所有语义写入都携带 `expectedRevision`，服务端在事务中校验版本、验证完整 Operation 批次并一次性提交；版本不匹配返回 HTTP 409，项目保持不变。视口保存属于体验状态，不增加语义 revision。

## Considered Options

- **最后写入覆盖**：实现简单，但会静默丢失用户或 Agent 的变更。
- **悲观锁**：可以阻止并发写入，但会增加连接占用和取消恢复复杂度；一期没有多人协作需求，不值得承担该成本。
- **完整操作日志合并**：适合实时协作，但需要冲突变换或 CRDT，明显超出一期范围。

## Consequences

- 客户端必须处理 409，刷新项目后重新生成或放弃当前提议。
- 事务边界必须包含 revision 检查和 Operation 应用，不能先更新 JSON 再异步更新版本。
- 未来引入多人协作时，可以在该 seam 上替换为操作合并协议，但一期数据模型不提前承担协作复杂度。
