# Keep the canvas renderer behind a domain-state seam

- Status: accepted
- Date: 2026-11-19

无限画布采用成熟渲染引擎（例如 React Flow 或同类方案）实现缩放、平移、选择和连线交互，但渲染引擎的节点状态不是业务事实来源。`features/canvas` 将引擎事件转换成 `Canvas Operation`，并把服务端项目快照映射回引擎模型。这样可以先利用成熟引擎交付交互，同时保留未来替换渲染器或支持自定义节点的空间。

## Considered Options

- **直接把渲染引擎 store 当作项目状态**：开发初期快，但会把第三方数据结构扩散到 Agent、持久化和测试，替换成本高。
- **从零实现画布引擎**：控制力更强，但会把大量时间投入到缩放、选择、连线和性能问题，不能服务一期核心闭环。

## Consequences

- 需要维护快照/Operation 与渲染器模型之间的适配器。
- 画布 reducer 和 Operation 转换可以脱离 DOM 做单元测试。
- 具体渲染库的选择仍需在实现阶段以 1,000 节点性能和可访问性验证为准，不改变业务接口。
