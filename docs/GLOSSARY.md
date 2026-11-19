# Glossary

## Canvas Project

用户可保存、打开和持续编辑的无限画布工作空间。一个项目包含节点、连线、视口和项目级元数据。

## Canvas Node

画布中的最小可操作对象，例如文本、图片、配置或 Agent 结果。节点拥有稳定标识、类型、位置和业务数据。

## Canvas Operation

对 Canvas Project 的一个结构化变更，例如新增节点、移动节点、建立连线或删除节点。用户操作和 Agent 操作都必须表达为 Canvas Operation。

## Revision

Canvas Project 的单调递增版本号，用于检测并发修改和避免旧数据覆盖新数据。

## Agent Session

围绕一个项目或创作目标持续存在的对话上下文，包含用户和 Agent 的消息历史。

## Agent Run

Agent Session 中一次可追踪、可取消、可重试的执行过程，从用户请求开始，到完成、失败或取消结束。

## Agent Step

Agent Run 中一个可审计的模型调用、Tool Call 或 Canvas Operation 执行步骤。

## Tool Call

Agent 请求执行一个具有明确输入约束和权限声明的工具。Tool Call 不允许绕过领域规则直接写数据库。

## Task

需要异步执行或轮询外部供应商的工作单元，例如媒体生成任务。Task 的状态必须持久化。

## Asset

项目或用户可引用的媒体及其元数据。大文件本体与画布 JSON 分离保存。
