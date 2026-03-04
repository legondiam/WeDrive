# WeDrive

基于 Go 的个人网盘后端服务，支持用户认证、文件上传与秒传、多级文件夹、回收站，使用 MinIO 作为对象存储。

---

## 业务亮点

- **秒传**：按文件内容哈希去重，相同文件只存一份，上传时若文件池中已有则只建用户关联，节省带宽与存储。
- **存储与成本优化**：`file_stores` 文件池 + 多用户 `user_files` 关联，天然去重，适合多用户共享同一文件的场景。
- **多级目录**：支持任意层级文件夹（无符号 `parent_id` 树形结构，根目录 `parent_id=0`），与常见网盘交互一致。
- **软删除与回收站**：文件/文件夹删除进入回收站，可恢复，避免误删；删除文件夹时会递归作用于整个目录树，列表接口只查未删除数据，逻辑清晰。
- **用户空间与引用计数**：按文件实际大小占用用户空间，永久删除时释放空间；通过引用计数只在无其他未删除引用时删除文件池记录与 MinIO 对象，避免重复存储。
- **无状态鉴权**：JWT 双 Token（Access + Refresh），刷新续期而不必频繁登录，便于多端使用。
- **文件夹零存储**：创建文件夹仅落库目录元数据（`IsFolder=true`，`file_store_id` 可空），不占对象存储空间。

---

## 功能概览

| 模块     | 能力 |
|----------|------|
| 用户     | 注册、登录、Token 刷新、JWT 鉴权 |
| 文件     | 上传、秒传、按目录列表、软删除、回收站列表与恢复、回收站永久删除（释放用户空间） |
| 文件夹   | 创建多级目录（仅元数据），与文件统一展示；删除文件夹时递归删除其下所有文件/子文件夹（软删除） |

技术栈：Go、Gin、GORM、MySQL、Redis、MinIO、JWT（配置见 `config/config.yaml`）。

---

## 项目结构

- `cmd/main.go`：入口，加载配置与 `app.Init()` 启动路由。
- `config/config.yaml`：端口、MySQL、Redis、MinIO、JWT 等配置。
- `internal/api`：用户与文件/文件夹 HTTP 接口。
- `internal/app`：依赖注入与应用初始化（如 wire）。
- `internal/initialize`：MySQL、Redis、MinIO 初始化。
- `internal/middleware`：JWT 鉴权中间件。
- `internal/model`：用户、文件池、用户文件模型。
- `internal/oss`：MinIO 封装。
- `internal/repository`：用户、用户缓存、文件仓储。
- `internal/router`：Gin 路由注册。
- `pkg/logger`、`pkg/utils`：日志与工具（hash、convert、jwts）。

---

## 配置说明

编辑 `config/config.yaml`，配置应用端口、MySQL、Redis、MinIO、JWT 等。

---

## API 概览

基础路径：`/api/v1`。

**公开接口**

- `POST /user/register` — 注册
- `POST /user/login` — 登录（返回 Access/Refresh Token）
- `POST /user/refresh` — 刷新 Access Token

**需登录（Header: `Authorization: Bearer <access_token>`）**

- `POST /file/upload` — 上传文件（FormData: `file`, `parent_id` 可选，支持秒传）
- `POST /file/upload-folder` — 创建文件夹（JSON: `name`, `parent_id`，默认 0 为根目录）
- `GET /file/list?parent_id=` — 当前目录下文件与文件夹列表
- `DELETE /file/delete/:ID` — 软删除到回收站；若为文件夹则递归删除其下所有文件/子文件夹
- `GET /file/recycle` — 回收站列表
- `POST /file/restore/:ID` — 从回收站恢复（文件或文件夹）
- `DELETE /file/permanent-delete/:ID` — 从回收站永久删除文件/文件夹：对普通文件会释放用户空间，并在无其他未删除引用时删除文件池记录与 MinIO 对象

---

## 开发说明

- 入口在 `cmd/main.go`，通过 `app.Init()` 完成配置、DB、Redis、MinIO、路由等初始化。
- 私有接口统一经 `AuthMiddleware()` 校验 JWT。
- 日志使用 zap，在 `pkg/logger` 中初始化，启动时 `logger.Init()`，退出前 `Sync()`。
