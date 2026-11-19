# 工程化基线

## 本地命令

```bash
cd web && pnpm install && pnpm check
cd ../backend && go test ./... && go vet ./...
```

根目录不是 Node 工程，不放置 `package.json` 或 Node lockfile。Next.js、Husky、Commitlint、lint-staged 和 Prettier 全部由 `web/package.json` 与 `web/pnpm-lock.yaml` 管理。不要在仓库中引入 npm、yarn 或其他重复 lockfile。

提交时由 Husky 自动执行：

- `pre-commit`: lint-staged，执行 ESLint 和 Prettier
- `commit-msg`: Commitlint，校验 Conventional Commits
- `pre-push`: 执行 `pnpm --dir web check`

## 模块边界

- `web/` 只负责浏览器交互、渲染和调用契约客户端。
- `backend/internal/httpapi` 只负责 HTTP 协议适配。
- `backend/internal` 中的 application/domain 模块不依赖 Gin。
- GORM 只能出现在 persistence adapter；禁止在启动流程调用 AutoMigrate。
- `contracts/openapi.yaml` 是 HTTP 契约的唯一来源。

## 环境变量

| 变量                       | 用途               | 默认                   |
| -------------------------- | ------------------ | ---------------------- |
| `APP_ENV`                  | 运行环境           | `development`          |
| `HTTP_ADDR`                | Go 服务监听地址    | `:8080`                |
| `DATABASE_URL`             | PostgreSQL 连接串  | 空（开发允许无库启动） |
| `NEXT_PUBLIC_API_BASE_URL` | 浏览器访问后端地址 | 无                     |

生产部署必须显式提供数据库连接、签名密钥和模型凭据；任何 secret 都不得进入前端变量或日志。
