# 工程化基线

## 本地命令

```bash
docker compose up -d postgres

cd backend && go run ./cmd/server -config config.yaml

cd ../web && pnpm dev
```

完整验证：

```bash
cd web && pnpm install && pnpm check
cd ../backend && go test ./... && go vet ./...
```

根目录不是 Node 工程，不放置 `package.json` 或 Node lockfile。Next.js、Husky、Commitlint 和 Prettier 全部由 `web/package.json` 与 `web/pnpm-lock.yaml` 管理。不要在仓库中引入 npm、yarn 或其他重复 lockfile。

提交时由 Husky 自动执行：

- `commit-msg`: Commitlint，仅校验 Conventional Commits
- `pre-push`: 执行 `pnpm --dir web check:push`，只做类型检查、测试和构建，不执行 ESLint

## 模块边界

- `web/` 只负责浏览器交互、渲染和调用契约客户端。
- `backend/internal` 采用 `controller → service → repository → model` 四层依赖方向。
- Controller 只负责 Gin、DTO 和 HTTP 错误映射；Service 承担业务规则和事务编排。
- Model 定义 GORM 数据库实体、字段映射和表名，不包含数据库执行逻辑或 HTTP DTO。
- Repository interface 隔离持久化执行逻辑；具体查询、写入和事务放在 `repository/gorm` adapter。
- Controller DTO 与 Model 分离；禁止在启动流程调用 AutoMigrate。
- `contracts/openapi.yaml` 是 HTTP 契约的唯一来源。

## 配置文件

Go 应用只从 YAML 配置读取运行参数：

- `backend/config.example.yaml`：提交到 Git，定义完整配置结构和安全默认值。
- `backend/config.yaml`：本机实际配置，包含数据库连接和模型密钥，被 Git 忽略。
- `-config`：启动参数，默认读取当前工作目录的 `config.yaml`。

首次配置：

```bash
cp backend/config.example.yaml backend/config.yaml
chmod 600 backend/config.yaml
```

本地开发时 Docker Compose 只启动 PostgreSQL，默认映射到 `127.0.0.1:5432`。Go 和 Next.js 均在 WSL 本地运行；Go 从 `backend/config.yaml` 读取数据库地址，Next.js 默认代理到 `127.0.0.1:8080`。

Backend/Web 的 Dockerfile仍保留用于镜像构建和部署，但不参与本地开发 Compose。

智谱模式使用以下内部模型映射：

- `canvas-agent` → `glm-4.7-flash`
- `canvas-image-fast` → `cogview-3-flash`
- `canvas-video-fast` → `cogvideox-flash`

在 `backend/config.yaml` 中填写密钥并把 `agent.provider` 设置为 `zhipu`。密钥不得写入 Compose、示例配置、源码或提交记录。

生产部署必须显式提供数据库连接、签名密钥和模型凭据；任何 secret 都不得进入前端变量或日志。
