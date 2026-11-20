# 无限画布一期需求评审稿

- 状态：已确认，进入规格设计
- 版本：0.1
- 日期：2026-11-19
- 目标：定义“Agent + 无限画布”第一期的可交付范围、领域接口和验收口径

## 1. 评审结论摘要

第一期建议交付一个“单用户、可持久化、可撤销、Agent 可提出并执行画布变更”的工作台，而不是一次性复刻参考项目的全部图片、视频、工作流和多供应商能力。

核心闭环为：

1. 创建或打开一个 Canvas Project。
2. 在无限画布上创建、编辑、移动、连接和删除文本节点。
3. 在 Agent 面板中输入目标，Agent 返回解释和结构化 Canvas Operation。
4. 用户预览并确认 Operation，系统以一次 Revision 原子提交。
5. 页面刷新后恢复项目、视口和最近一次 Agent 会话。

一期明确不承诺多人实时协作、复杂媒体供应商编排、视频/音频生成、插件市场和 Python 运行时。图片节点可以预留类型，但不作为一期验收前提。

## 2. 参考项目的取舍

参考 `/home/yanbin/projects/infinite-canvas` 的有效经验：

- `CanvasProject`、`CanvasNode`、`CanvasConnection` 的概念已经比较稳定。
- 画布 JSON 与媒体 Blob 分离，避免把大文件直接塞入项目文档。
- 生成任务、Agent 会话和资产有独立状态，适合后续扩展。
- 节点操作手册和后端数据结构文档可以作为产品行为说明的模板。

一期不直接照搬的部分：

- 多种图片/视频/音频供应商和 ComfyUI 等外部编排。
- 浏览器 `localForage` 作为唯一持久化层。
- 将所有生成参数、供应商协议和媒体兼容逻辑放进画布节点。
- 以大量 handler/service/repository 暴露细粒度接口，导致前后端接口和领域规则重复。

取舍原则：画布只保存可编辑的领域状态；外部模型、异步任务和媒体资产通过独立模块接入，不能绕过 Canvas Operation 直接改画布。

## 3. 一期用户故事与范围

### 3.1 必须交付（P0）

| 编号  | 用户故事                                     | 验收结果                                              |
| ----- | -------------------------------------------- | ----------------------------------------------------- |
| P0-01 | 用户可以创建、重命名、打开和删除项目         | 项目列表和项目详情可用，刷新后数据不丢失              |
| P0-02 | 用户可以平移和缩放无限画布                   | 缩放中心稳定，节点位置使用世界坐标，刷新后恢复视口    |
| P0-03 | 用户可以新增、编辑、移动、框选和删除文本节点 | 节点操作不改变其它节点；删除节点会清理相关边          |
| P0-04 | 用户可以在两个节点之间建立和删除有向连接     | 不允许连接不存在节点；项目不能产生无效边              |
| P0-05 | 用户可以撤销和重做最近的画布变更             | 至少覆盖节点和连接操作，刷新后历史可丢弃              |
| P0-06 | 用户可以向 Agent 发送问题或目标              | Agent Run 有状态、消息可见、失败可重试或取消          |
| P0-07 | Agent 只能通过结构化 Operation 提议变更      | 用户确认前不写入项目；确认后一次性提交并增加 revision |
| P0-08 | 页面可以恢复项目和最近会话                   | 项目 JSON、revision、视口和会话消息可恢复             |

### 3.2 建议交付（P1）

- 图片节点：上传一张图片并作为 Agent 上下文引用，但先不接图片生成供应商。
- Operation 预览：在画布上高亮将新增、修改和删除的节点。
- 自动保存状态提示、离线草稿和最近项目。
- Agent 生成文本节点和连接的快捷指令。

### 3.3 明确排除（一期不验收）

- 多人实时协作、在线 presence、冲突合并和权限分享。
- 视频/音频生成、工作流节点、供应商协议适配和计费。
- 任意代码执行、浏览器自动化、文件系统写入等高权限 Agent 工具。
- 以 Python 作为产品请求链路的必要依赖。

## 4. 推荐领域模型

### 4.1 Canvas Project

```ts
type CanvasProject = {
  id: string;
  title: string;
  revision: number;
  nodes: CanvasNode[];
  edges: CanvasEdge[];
  viewport: { x: number; y: number; zoom: number };
  createdAt: string;
  updatedAt: string;
};
```

一期节点只要求 `text` 类型，但 `type` 必须是可扩展枚举。`viewport` 是用户体验状态，不应参与 Agent 的语义变更判断。

```ts
type CanvasNode = {
  id: string;
  type: "text" | "image";
  title: string;
  position: { x: number; y: number };
  size: { width: number; height: number };
  data: { text?: string; assetId?: string };
};

type CanvasEdge = {
  id: string;
  source: string;
  target: string;
};
```

### 4.2 Canvas Operation

Operation 是唯一的画布写入语言。前端手势、快捷键和 Agent 都应先转换成 Operation，再由同一个领域模块校验和应用。

一期最小集合：

- `node.create`
- `node.update`
- `node.move`
- `node.delete`
- `edge.create`
- `edge.delete`

应用规则：整批 Operation 使用 `expectedRevision` 做乐观并发校验；全部通过才提交，失败则不产生部分写入。服务端返回新 `revision` 和规范化后的项目摘要。

### 4.3 Agent Run

Agent Run 至少包含 `queued/running/succeeded/failed/cancelled` 状态、输入消息、输出消息、结构化 Operation、错误信息和时间戳。模型输出必须经过 schema 校验、权限校验和领域校验；模型不能直接调用 GORM 或修改项目表。

## 5. 推荐模块与接口

### 5.1 Web 模块

```text
web/
  app/                  # 路由和页面组合
  features/canvas/      # 画布渲染、手势、选择和操作转换
  features/agent/       # 会话、流式事件、Operation 预览和确认
  features/projects/    # 项目列表、打开、保存
  domain/canvas/        # 前端可复用类型和纯操作 reducer
  lib/api/              # OpenAPI 客户端适配
```

`features/canvas` 对外只暴露一个深模块接口：输入项目快照和用户意图，输出渲染状态或 `CanvasOperation[]`。画布库（React Flow、Konva 或其它实现）不能成为业务状态的唯一来源。

### 5.2 Backend 模块

```text
backend/internal/
  controller/           # Gin 路由、DTO、SSE 和 HTTP 错误映射
  service/              # 项目、Canvas Operation 和 Agent Run 用例
  repository/           # 持久化 interface、具体查询、GORM 和 Memory adapter
  model/                # GORM 实体结构、字段关系和表名映射
  platform/             # 数据库连接等共享基础设施
```

一期采用 Go Agent runtime。Python 只在以后需要 OCR、CV、embedding 或 GPU 推理时作为独立 worker 接入；worker 只能返回结构化结果，不能直接提交画布变更。

## 6. 一期 HTTP 契约草案

最终以 `contracts/openapi.yaml` 为准，建议补齐以下资源：

| 方法     | 路径                                   | 作用                      |
| -------- | -------------------------------------- | ------------------------- |
| `GET`    | `/projects`                            | 项目列表                  |
| `POST`   | `/projects`                            | 创建项目                  |
| `GET`    | `/projects/{projectId}`                | 获取项目快照              |
| `PATCH`  | `/projects/{projectId}`                | 修改标题或视口            |
| `DELETE` | `/projects/{projectId}`                | 删除项目                  |
| `POST`   | `/projects/{projectId}/operations`     | 校验并提交 Operation 批次 |
| `GET`    | `/projects/{projectId}/agent/sessions` | 获取会话                  |
| `POST`   | `/projects/{projectId}/agent/runs`     | 创建 Agent Run            |
| `GET`    | `/agent/runs/{runId}/events`           | SSE 流式事件              |
| `POST`   | `/agent/runs/{runId}/cancel`           | 取消运行                  |

`POST /operations` 必须支持 `expectedRevision`，并在 revision 不匹配时返回明确的冲突错误（建议 HTTP 409）。Agent SSE 事件至少定义 `run.started`、`message.delta`、`operation.proposed`、`run.completed`、`run.failed`。

## 7. 非功能验收标准

- 首次打开已有项目时，画布主界面在常规开发环境下可交互；大项目性能指标先以 1,000 个节点为测试上限。
- 所有写操作具备幂等键或客户端请求 ID，重复提交不能重复创建节点。
- 任意非法 Operation 都返回可定位的字段级错误，不允许静默忽略。
- Agent 输出不符合 schema、超出节点数量限制或引用不存在节点时，项目保持原 revision。
- 断开 SSE 后可以查询 Run 当前状态；取消后不能继续提交 Operation。
- 画布 JSON 不内嵌大文件；资产使用独立 `assetId`，具体存储实现不进入一期验收。
- 单元测试覆盖 Operation reducer/validator、revision 冲突、Agent 输出解析和 SSE 状态转换；端到端测试覆盖“创建项目—Agent 提议—确认提交—刷新恢复”。

## 8. 需要产品确认的决策

以下决定已确认，作为一期规格设计的约束：

1. **一期节点类型**：只验收文本节点，图片节点仅预留或支持基础上传。
2. **Agent 写入模式**：先预览，用户确认后提交；不提供默认自动执行。
3. **项目身份**：只做单用户私有项目，暂不做登录、分享和权限。
4. **保存策略**：服务端持久化为主，浏览器只做临时草稿。
5. **流式体验**：采用 SSE 流式输出，并支持断线后查询 Run 状态。
6. **画布引擎**：允许使用 React Flow 等成熟引擎，业务状态仍由 `CanvasOperation` 管理。
7. **连接语义**：一期只表达上下文/引用关系，暂不区分端口、数据类型和执行顺序。

## 9. 当前实现状态

- `contracts/openapi.yaml` 已覆盖项目 CRUD、Canvas Operation 批次和 revision 冲突；Agent Run 与 SSE 契约待下一切片补充。
- `web/app/page.tsx` 已替换为项目入口，并新增基于 React Flow adapter 的画布工作台。
- `backend/internal` 已按四层架构实现项目和 Operation 接口，业务规则位于 Service。
- 旧的根目录 `server/main.go` 已删除，Go 应用唯一入口为 `backend/cmd/server/main.go`。

## 10. 评审出口

本稿评审通过的标准：第 8 节的 7 个决策全部有结论；P0 用户故事每项都有对应接口、状态变化和验收测试；OpenAPI、Glossary、数据库迁移和前端状态模型使用同一组术语。
