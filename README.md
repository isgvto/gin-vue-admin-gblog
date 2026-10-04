# gin-vue-admin-gblog

基于 [gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin) 改造的个人博客系统，包含 Go 后端、Vue 3 管理后台和独立的 Vue 2 博客前台。博客业务复用 GVA 的用户、角色、菜单、JWT 登录与 Casbin 权限体系，支持 Markdown 写作、AI 辅助编辑、配图、访客统计与 GitHub 文档阅读。

当前开发与运行入口为 **`server`、`web`、`blog-view`**。

## 导航

- [功能概览](#功能概览)
- [项目结构](#项目结构)
- [环境要求](#环境要求)
- [本地启动](#本地启动)
- [可选功能配置](#可选功能配置)
- [构建与测试](#构建与测试)
- [部署说明](#部署说明)
- [接口与开发入口](#接口与开发入口)
- [常见问题](#常见问题)
- [许可证与致谢](#许可证与致谢)

## 功能概览

| 模块 | 当前能力 |
| --- | --- |
| 博客前台 | 文章列表与详情、置顶与推荐、分类与标签、归档与搜索、密码文章、评论与回复、关于页、友链、动态与点赞 |
| 内容管理 | Markdown 编辑、发布与可见性、评论与赞赏开关、分类与标签管理、评论审核、友链与动态管理、站点设置 |
| AI 写作 | 多轮写作对话，支持润色、改写、续写、审阅、大纲、标题、摘要与标签建议；浏览器本地会话、差异预览、逐块采纳、撤销与按章节写作 |
| 配图与图示 | AI 配图方案与图片生成、预览后采用或丢弃；编辑器、文章及文档阅读页支持 Mermaid 图示 |
| 文档与集成 | GitHub 文档目录与内容读取、手动同步及 Webhook；独立 MCP 服务、AI 模型管理 |
| 统计与运维 | 博客仪表盘、阅读量、访问日志、访客统计、操作日志与异常日志 |
| GVA 基础能力 | 用户、角色、菜单、API、权限、文件上传、字典、代码生成、系统工具及插件 |

AI、图片生成、对象存储和 GitHub 文档需要各自的配置；基础博客浏览与内容管理不要求提供模型密钥。

## 项目结构

```text
.
├── server/                 # Gin + GORM 后端（GVA、博客与 AI 服务）
│   ├── api/v1/blog/        # 博客 HTTP 接口与流式响应
│   ├── service/blog/       # 博客业务逻辑
│   ├── service/ai/         # 模型工厂及 AI 集成
│   ├── model/blog/         # 博客模型与请求结构
│   ├── router/blog/        # 公开接口与后台接口路由
│   ├── initialize/         # 数据库迁移、路由及权限升级
│   ├── source/             # 首次初始化数据（用户、菜单、权限、站点设置等）
│   └── cmd/mcp/            # 独立 MCP 服务入口与配置
├── web/                    # Vue 3 + Vite 管理后台
│   ├── src/api/blog/       # 博客接口封装
│   ├── src/view/blog/      # 文章、分类、标签、评论、动态与仪表盘
│   └── tests/ai-editor/    # AI 编辑器回归与浏览器测试夹具
├── blog-view/              # Vue 2 + Vue CLI 博客前台
├── docs/                   # 仓库说明资源，非在线文档内容源
├── deploy/                 # 历史 Docker / Compose / Kubernetes 部署模板
├── docker-compose.yml      # 三个应用的预构建镜像编排
└── Makefile                # GVA 构建、Swagger 与插件打包任务
```

## 环境要求

以下版本以仓库中的依赖声明为准：

| 依赖 | 要求与用途 |
| --- | --- |
| Go | `server/go.mod` 声明 Go `1.24.0`、工具链 `go1.24.2` |
| Node.js / npm | 统一本地环境可使用 Node.js **24+**、npm **10+**，与 `blog-view/package.json` 的 `engines` 一致；前台 Dockerfile 也使用 Node 24 |
| 数据库 | 下文使用 MySQL，需支持 `utf8mb4` 并使用 InnoDB；GVA 连接层还包含 PostgreSQL、SQLite、SQL Server、Oracle，博客功能在这些数据库上的兼容性需单独验证 |
| Redis | `system.use-redis: true` 时需要；多点登录和启用每日限额的 AI 生成也依赖相关 Redis 配置 |
| Docker / Compose | 仅容器部署需要 |

后台使用 Vue 3、Vite 6、Element Plus、Pinia、Axios、ECharts；前台使用 Vue 2、Vue CLI 5、Vuex、Vue Router、Element UI 与 Semantic UI。后端使用 Gin、GORM、JWT、Casbin、Zap、Viper、Goldmark 和 Eino。

## 本地启动

以下命令均从仓库根目录开始，三个应用分别在独立终端启动。

### 1. 准备后端配置

复制完整配置作为本机配置：

```powershell
cd server
Copy-Item config.yaml config.local.yaml
```

Linux / macOS 可用 `cp config.yaml config.local.yaml`。编辑 `server/config.local.yaml` 中对应字段，保留其余配置。下面是本地开发示例，**不是完整配置文件**：

```yaml
mysql:
  path: 127.0.0.1
  port: "3306"
  username: root
  password: "替换为本机数据库密码"
  db-name: ""               # 首次安装留空，在后台初始化页面创建数据库
  config: charset=utf8mb4&parseTime=True&loc=Local
  engine: InnoDB

redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0

system:
  db-type: mysql
  addr: 8888
  router-prefix: ""
  oss-type: local           # 本地上传；使用云存储时需另行填写凭据
  use-redis: true
  use-multipoint: false
  disable-auto-migrate: false
```

同时替换 `jwt.signing-key`。复制的配置仍可能包含原有数据库地址、云存储和站点参数，启动前应按本机环境检查；不要提交实际密码、密钥或含凭据的本机配置文件。

已有初始化完成的数据库可直接填写 `mysql.db-name` 及连接信息。首次安装请保持库名为空：后端据此进入未初始化状态，而不是尝试连接已有数据库。

### 2. 启动后端

在上一步的 `server` 终端执行：

```sh
go mod download
go run . -c config.local.yaml
```

配置选择顺序为 `-c` 参数、`GVA_CONFIG` 环境变量、Gin 模式对应配置文件，找不到模式配置文件时回退到 `config.yaml`。从 `server` 目录运行可确保资源与配置的相对路径正确。

| 地址 | 用途 |
| --- | --- |
| `http://127.0.0.1:8888/health` | HTTP 健康检查，返回 `"ok"`；不代表数据库已初始化 |
| `http://127.0.0.1:8888/swagger/index.html` | 已生成的 Swagger 文档 |

### 3. 启动管理后台并初始化

在新的根目录终端执行：

```sh
cd web
npm ci
npm run dev
```

访问 `http://127.0.0.1:8080`。`npm run serve` 与 `npm run dev` 等价，均启动 Vite 开发服务。

首次安装在后台初始化页面填写数据库地址、账号、密码、库名（例如 `gva`）和管理员密码。数据库账号需要有建库、建表权限。初始化会写入用户、角色、菜单、API、权限和站点基础数据，并回写当前使用的配置文件；因此本机配置文件需要可写。完成后重启后端，使用 `admin` 和初始化时设置的密码登录。

**自动迁移与首次初始化用途不同**：`disable-auto-migrate: false` 在连接到数据库后迁移系统表、博客表与 AI 表，但不会代替首次初始化创建完整的账号、菜单与权限种子数据。不要仅创建空数据库后就直接尝试登录。

后台开发代理读取 [web/.env.development](web/.env.development)：

```env
VITE_CLI_PORT = 8080
VITE_SERVER_PORT = 8888
VITE_BASE_API = /api
VITE_FILE_API = /api
VITE_BASE_PATH = http://127.0.0.1
```

### 4. 启动博客前台

在另一个根目录终端执行：

```sh
cd blog-view
npm ci
npm run serve -- --port 8081
```

访问 `http://127.0.0.1:8081`。显式指定端口可避免与管理后台的 `8080` 冲突；`vue.config.js` 本身没有固定开发端口。

两个前端均使用 `/api` 请求前缀，开发代理会移除此前缀后转发到 `http://127.0.0.1:8888`。例如浏览器请求 `/api/blogs`，后端收到 `/blogs`。调整后端地址时同步修改后台环境配置和 [blog-view/vue.config.js](blog-view/vue.config.js) 的代理目标。

登录后台后，可先设置站点信息、分类与标签，再发布一篇文章验证前台展示。

## 可选功能配置

### AI 写作与配图

在后台「AI 与集成 → AI 模型配置」中新增模型，填写供应商、接口地址、API Key 和模型标识，启用并设为默认，测试连接后再使用写作助手。模型工厂支持 `openai`、`ark`、`gemini`，数据库中的默认启用模型优先于 YAML `ai` 配置；YAML 仅在未找到数据库默认模型且 `enable` 为真、密钥非空时兜底。

`ai.context-limit` 控制正文取材预算，`ai.daily-limit` 控制每日生成次数。启用每日限额时，Redis 不可用会拒绝生成；模型失败或用户取消不会退回已经计入的次数。长文摘要使用抽样窗口，并不保证覆盖全文所有细节。

文字助手通过同一对话框理解写作要求，结果由作者选择对比、插入或回填。聊天记录和未发送输入按用户及文章保存在当前浏览器 `localStorage`，不建立聊天数据库；可新建、切换、删除或清空对话；删除会话会释放对应的本地存储空间。每篇文章最多保留最近 5 个会话，每个会话最多 40 条消息，超过容量会淘汰旧记录。刷新后只有原文章正文仍一致时才允许应用旧结果。浏览器数据清理后无法恢复，记录不跨设备同步。图示配图继续使用独立的预览、采用与上传流程。自动流程图保留关键步骤、分支及反馈关系，允许最多4个真实阶段分组（不嵌套）；24个节点、40条连线与36字标签仅作为安全上限。渲染后按720px参考正文宽度与420px预览高度估算缩放字号、检查长宽比，必要时尝试TD/LR布局，保持节点与连线不变；若仍难阅读则提示拆分，作者可准备分阶段要求并确认重新生成。原始源码保留，布局优化也用于高清导出。后台与博客前台的 Mermaid 正文预览高度最多420px，可打开完整图示缩放查看，高清导出分辨率不受预览尺寸影响。

快捷操作会合并输入框中的补充要求；审阅建议可定位原文、讨论或按建议修改单段。对比中的采用与保留立即生效，后续追问引用实际保留的正文。选择旧回复继续调整时会显示引用范围；阅读聊天历史时暂停自动滚动，可点击「查看最新回复」返回底部。

图片生成还需要配置图片模型及存储。配图先展示预览，由作者确认采用后上传；若使用云存储，请填写 `system.oss-type` 对应的存储配置。AI 生成结果不会自动发布文章。

### GitHub 文档

在线文档通过站点设置中的 `docsGithubRepo`、`docsGithubBranch`、`docsGithubRoot` 指定来源，支持公开的 `/docs/tree`、`/docs/content` 和后台 `/admin/docs/sync`。使用 `/docs/webhook` 同步时还需设置 `docsGithubWebhookSecret`。仓库根目录的 `docs/` 不是该功能自动读取的内容源。

### 独立 MCP 服务

MCP 不随博客后端自动启动。需要时在 `server` 目录另开终端：

```sh
go run ./cmd/mcp -config ./cmd/mcp/config.yaml
```

默认地址为 `http://127.0.0.1:8889/mcp`，配置中的 `upstream_base_url` 指向博客后端，默认认证头为 `x-token`。修改服务地址或部署路径时检查 [MCP 配置](server/cmd/mcp/config.yaml)。

## 构建与测试

在各模块目录分别执行：

| 模块 | 命令 | 结果或用途 |
| --- | --- | --- |
| `server` | `go build -o gblog-server .` | 构建后端二进制；Windows 可指定 `gblog-server.exe` |
| `server` | `go test ./...` | 后端测试；部分集成用例依赖外部服务或环境变量 |
| `web` | `npm run build` | 输出 `web/dist` |
| `web` | `npm run test:ai` | AI 编辑器纯逻辑回归测试 |
| `web` | `npm run test:ai:serve` | 启动独立浏览器测试夹具，默认端口 `5191` |
| `blog-view` | `npm run build` | 输出 `blog-view/dist` |

AI 相关后端可单独验证：

```sh
cd server
go test ./service/blog ./service/ai ./api/v1/blog ./api/v1/ai
```

浏览器回归和专用 Redis 配额测试的准备方式见 [AI 编辑器测试说明](web/tests/ai-editor/README.md)。配额集成测试通过 `AI_TEST_REDIS_ADDR` 指定临时 Redis，使用 DB 15，不能指向业务 Redis。

后端运行时仍需配置、`resource/` 和上传等相关目录；仅复制二进制不足以提供完整运行环境。

## 部署说明

### 根目录 Docker Compose

[docker-compose.yml](docker-compose.yml) 使用指定版本的预构建镜像：

| 服务 | 镜像 | 宿主机端口 |
| --- | --- | --- |
| 后端 | `gvto/gva-server:v4.1.0` | `8888` |
| 管理后台 | `gvto/gva-web:v4.1.0` | `8080 → 80` |
| 博客前台 | `gvto/blog-view:v4.0.2` | `8081 → 80` |

使用前需完成以下配置：

1. 提供可从容器访问的数据库与 Redis；根目录 Compose 没有定义这两个服务，容器中的 `127.0.0.1` 也不是宿主机。
2. 为后端挂载或提供自己的完整配置。当前 Compose 未挂载配置文件，直接启动会使用镜像内置配置；挂载路径需与实际镜像工作目录一致。
3. 核对日志、上传与运行缓存卷的路径。当前 Compose 使用 `github.com/LZMclear/...` 路径，源码 Dockerfile 使用 `github.com/isgvto/...` 工作目录，自建镜像时应同步调整。
4. 确认初始化与存储配置完成，再启动应用。

完成上述准备后，在仓库根目录执行：

```sh
docker compose up -d
docker compose ps
docker compose logs -f gva-gbserver
```

修改源码后，此命令不会自动构建新镜像。若需要部署当前代码，请分别以 `server/`、`web/`、`blog-view/` 为构建上下文构建镜像，再更新 Compose 镜像引用。前台 Nginx 的上游可由 `API_PROXY_PASS` 设置，默认 `http://gva-gbserver:8888`；后台 Nginx 配置也使用此服务名。

### 静态资源与反向代理

两个前端构建产物分别部署到各自的静态站点。生产 Nginx 需要将 `/api/` 转发到后端并移除代理前缀，同时为前端路由设置 `try_files ... /index.html`。AI 流式接口还需按部署环境配置代理缓冲与读取超时，避免输出被集中缓冲或提前断开。

后端 `system.router-prefix` 默认为空。如果改为非空，健康检查、Swagger 和业务接口都会增加此前缀，前端代理与权限配置也需要对应检查。

### 历史部署模板

`deploy/` 和 `Makefile` 保留了 GVA 部署模板，使用前需要校准版本、端口、服务名和路径。例如 Makefile 的 Go 构建镜像仍为 `golang:1.22`，低于当前模块要求，且 `make build` 未覆盖独立博客前台；`deploy/docker-compose/docker-compose.yaml` 的后台端口映射和上游服务名也与当前 Nginx 配置不一致。它们不应直接作为当前三应用部署的默认入口。

Swagger 可在安装 `swag` CLI 后通过根目录 `make doc` 更新；Makefile 依赖 Bash，Windows 可在具备相应工具的 WSL / Bash 环境使用。

## 接口与开发入口

以下路径以 `system.router-prefix: ""` 为前提。公开接口无需后台 JWT；私有接口继承 GVA JWT 与 Casbin，通常通过 `x-token` 传递登录令牌。

| 路由组 | 示例 | 权限 |
| --- | --- | --- |
| 站点与文章 | `GET /site`、`GET /blogs`、`GET /blog`、`GET /searchBlog`、`GET /archives` | 公开 |
| 分类与标签 | `GET /category`、`GET /category/blogs`、`GET /tag`、`GET /tag/blogs` | 公开 |
| 评论与页面 | `GET /comments`、`POST /comment`、`GET /about`、`GET /friends`、`GET /moments` | 公开 |
| 密码与互动 | `POST /checkBlogPassword`、`POST /friend`、`POST /moment/like/:id` | 公开，业务层校验 |
| GitHub 文档 | `GET /docs/tree`、`GET /docs/content`、`POST /docs/webhook` | 公开；Webhook 验证签名 |
| 博客管理 | `/admin/blogs`、`/admin/blog`、`/admin/categories`、`/admin/tags`、`/admin/comments` | JWT + Casbin |
| 配置与统计 | `/admin/siteSettings`、`/admin/dashboard`、`/admin/visitLogs`、`/admin/visitors` | JWT + Casbin |
| AI 写作 | `GET /blog/ai/status`、`POST /blog/ai/chat`、`POST /blog/ai/summary`、`POST /blog/ai/suggest-tags` | JWT + Casbin |
| AI 配图 | `GET /blog/ai/visual/status`；`POST /blog/ai/visual/plan`、`POST /blog/ai/visual/generate`、`POST /blog/ai/visual/adopt`、`POST /blog/ai/visual/discard` | JWT + Casbin |

实际方法、参数和路径以 [博客路由](server/router/blog)、[API 实现](server/api/v1/blog) 与请求模型为准，Swagger 只反映已经生成的注解内容。AI 接口位于 `/blog/ai`，并非 `/admin`。

开发时可沿下列路径定位代码：

- 后端：`router/blog → api/v1/blog → service/blog → model/blog`；路由注册入口为 [router_biz.go](server/initialize/router_biz.go)。
- 数据库：启动迁移入口为 [gorm.go](server/initialize/gorm.go) 与 [gorm_biz.go](server/initialize/gorm_biz.go)，首次初始化数据在 `server/source/`。主要业务模型包括文章、分类、标签、评论、友链、动态、站点设置、访问与操作日志及 AI 配图任务。
- 后台：接口在 `web/src/api/blog`，页面在 `web/src/view/blog`、`blogPage`、`blogLog`、`blogStatistics`，AI 配置页面在 `web/src/view/ai`。
- 前台：接口在 `blog-view/src/api`，页面在 `blog-view/src/views`。文章 Markdown 由后端渲染为 HTML，Mermaid 在前端按需渲染；修改共享图示渲染逻辑时检查两个前端副本。
- 权限：新增后台接口时同步检查 API 注册、菜单与 Casbin 授权，仅增加路由不能保证角色拥有访问权限。

## 常见问题

### 后端启动时数据库连接失败

检查实际选中的配置文件，以及数据库地址、端口、账号和库名。首次安装使用空 `db-name` 进入初始化流程；非空库名连接失败会导致启动报错，而不会自动进入初始化。已有数据库请确认网络与凭据，勿用重新初始化代替排查。

### 数据表已创建，但没有账号或菜单

自动迁移只负责表结构。确认是否完成后台初始化、是否使用正确的数据库，以及用户的菜单与 API 权限。`/init/checkdb` 当前依据数据库连接是否存在判断初始化状态，不检查种子数据是否完整。

### 前端请求失败或后台返回权限错误

先验证 `/health`，再查看浏览器 Network 的请求路径和返回体。检查 `/api` 代理目标、后端路由前缀、登录令牌与角色 API 权限。修改开发环境变量后重启前端服务。GVA 业务错误可能仍使用 HTTP 200，需同时检查响应中的 `code` 与 `msg`。

### 上传失败或 AI 配图无法采用

检查 `system.oss-type` 对应的凭据、桶地址或本地存储路径，以及容器卷权限。仓库配置选择的云存储不一定在本机可用，本地开发可改用 `local` 并保留完整的 `local` 配置。

### AI 不可用或输出中断

检查默认启用模型、密钥、供应商地址与模型标识；每日限额开启时还需检查 Redis。通过代理部署时检查 SSE 缓冲和超时。更完整的配额、取消和重试行为见 [AI 编辑器测试说明](web/tests/ai-editor/README.md)。

### 博客前台端口冲突或依赖报错

使用 `npm run serve -- --port 8081` 避免占用后台端口。依赖问题先核对 Node.js 24+、npm 10+ 与锁文件，用 `npm ci` 安装；当前前台已使用 Vue CLI 5，不需要默认添加旧版 Webpack 的 OpenSSL 兼容参数。

### 出现 `/sysError/createSysError` 请求

这是后台运行时错误上报。先查看浏览器控制台中的原始错误，检查字段、组件或接口响应结构，再处理上报请求本身。

## 许可证与致谢

本仓库采用 [Apache License 2.0](LICENSE)。项目基于 gin-vue-admin 改造，使用和分发时请保留原有版权与许可证声明；第三方依赖及资源遵循各自的许可证。
