# WeDrive

WeDrive 是一个前后端分离的个人网盘项目。后端使用 Go/Gin/GORM，文件对象存储在 MinIO，MySQL 保存业务元数据，Redis 保存上传过程中的临时状态；前端使用 Vue 3 + Vite + Pinia + FilePond 实现文件管理、上传、回收站和分享下载。

这个项目的重点不是简单 CRUD，而是围绕“文件上传链路”做了分层秒传、对象存储直传、分块续传、文件池去重和用户空间一致性处理。

---

## 技术栈

**后端**

- Go、Gin、GORM
- MySQL：用户、文件池、用户文件、上传会话、分享记录等业务数据
- Redis：秒传挑战会话、分块上传 ETag/Hash 临时状态
- MinIO：对象存储、预签名下载、预签名分块上传
- JWT：Access Token / Refresh Token 鉴权

**前端**

- Vue 3、Vite、Pinia、Vue Router
- FilePond：上传交互
- Web Crypto API：前端计算文件 hash、抽样 hash、分块 hash

---

## 核心设计

### 1. 分层秒传链路：QuickCheck + Full Hash + PoE

秒传不是简单地“客户端传 hash，服务端命中就创建文件记录”。项目中把秒传拆成三层：

```text
QuickCheck 抽样预筛
-> PrepareInstantUpload 完整 hash 命中并生成随机挑战
-> InstantUpload 校验随机片段证明并落库
```

**第一层：QuickCheck 抽样预筛**

前端先计算：

```text
file_size
head_hash
mid_hash
tail_hash
```

然后调用：

```text
POST /api/v1/file/quick-check
```

后端通过 `file_size + head_hash + mid_hash + tail_hash` 查询 `file_stores`，只判断“是否有秒传可能”。如果抽样都没命中，前端不再计算完整文件 hash，直接进入普通上传或分块上传，避免大文件无意义地全量读盘。

**第二层：PrepareInstantUpload 完整 hash 精确命中**

QuickCheck 命中后，前端再计算完整 SHA-256，并调用：

```text
POST /api/v1/file/instant-upload/prepare
```

后端用 `hash_type + file_hash` 精确查询文件池。如果命中，服务端生成若干随机片段挑战：

```json
[
  { "offset": 1024, "length": 4096 },
  { "offset": 88888, "length": 4096 }
]
```

并把 `prepare_id` 对应的挑战会话写入 Redis，状态中绑定：

```text
user_id
file_store_id
parent_id
hash_type
file_hash
challenges
```

这样挑战由服务端生成，客户端不能自己挑片段，也不能把别人的挑战拿来复用。

**第三层：InstantUpload 随机片段所有权证明**

前端根据 challenges 从本地文件读取对应字节片段，base64 后提交：

```text
POST /api/v1/file/instant-upload
```

后端会：

1. 读取 Redis 中的 `prepareState`。
2. 校验 `prepare_id` 是否存在、是否属于当前用户。
3. 校验当前请求和 prepare 阶段的 `parent_id/hash_type/file_hash` 是否一致。
4. 校验客户端返回的 `offset/length` 是否确实是服务端挑战过的片段。
5. 从 MinIO 按相同 offset/length 读取已存文件的真实字节区间。
6. 使用 `bytes.Equal` 对比客户端片段和服务端片段。
7. 校验通过后删除挑战会话，并创建当前用户的文件记录。

这条链路解决的是“只知道热门文件 hash 就伪造秒传”的问题。完整 hash 只证明文件身份，随机片段 PoE 才进一步证明客户端大概率真实持有该文件。

---

### 2. MinIO 分块直传与断点续传

大文件不会先上传到业务服务器再转存，而是走 MinIO 分块直传：

```text
InitChunkUpload
-> SignPartUpload
-> 前端 PUT 到 MinIO 预签名 URL
-> ReportUploadedPart
-> CompleteChunkUpload
```

核心设计：

- 后端创建 MinIO multipart upload，并在 MySQL 中保存上传会话。
- 前端按固定分块大小计算每个 chunk 的 SHA-256。
- `SignPartUpload` 为单个分块生成 MinIO 预签名上传 URL，并把该分块的 SHA-256 checksum 放入签名请求中。
- 分块上传 URL 和 `chunk_hash` 绑定后，前端实际 PUT 到 MinIO 的分块内容必须和签名时声明的 hash 一致，避免客户端向后端报告一个 hash、实际上传另一个内容。
- 预签名 URL 中携带签名字段，请求会经过网关层；可以在 Nginx 按签名上传路径/请求特征配置限速，把大文件上传的流量控制放在网关侧，而不是压到业务服务上。
- 前端上传成功后回报 ETag，Redis 记录已上传分块。
- 重新初始化同一文件上传时，服务端返回已上传分块编号，前端跳过这些分块，实现断点续传。
- `CompleteChunkUpload` 会检查 ETag 数量和顺序，最后调用 MinIO complete multipart upload。

这条链路把大文件流量从业务服务中剥离出去，业务服务只负责签名、状态管理和最终一致性处理；分块内容校验交给对象存储 checksum，上传流量治理交给网关层。

---

### 3. 多层限流与带宽治理

项目没有把限流只做成一个全局 QPS 开关，而是按“数据流量”和“业务控制面”分层治理：

- **网关层管大流量**：大文件分块内容通过 MinIO 预签名 URL 直传，不经过业务服务；上传流量可以在 Nginx 按预签名上传路径/请求特征做限速。下载 URL 会被包装成 `/WeDrive/{exp}/{tier}/{signature}/{object}` 形式，签名内容绑定 `uri + exp + tier + secret`，网关校验 secure link 后可以按 `tier` 配置不同下载限速策略。
- **业务层管控制接口**：`init/sign-part/report-part/complete/quick-check/instant-*` 这些小 JSON 接口走 Redis 令牌桶限流，按 `user_id`、`upload_id` 做不同粒度控制，防止刷签名、刷 Redis 状态和高频探测文件 hash。
- **资源层限制堆积**：`InitChunkUpload` 除了限频，还限制单个用户未完成的 pending 上传会话数，避免恶意用户慢速创建大量 multipart upload，占住 MySQL/MinIO 清理资源。
- **最终提交做幂等保护**：`CompleteChunkUpload` 是上传链路的最终提交动作，不只依赖限流，还通过幂等状态避免重复提交造成假失败或重复落库。

这套设计的边界比较清楚：Nginx 负责上传/下载字节流的带宽治理，Redis 令牌桶负责业务控制接口的请求速率，数据库状态和上传会话上限负责资源占用约束。

---

### 4. 文件池去重、用户空间与对象生命周期

项目把“物理文件”和“用户文件”拆成两张表：

```text
file_stores：物理文件池，保存 hash、size、MinIO object name、抽样 hash
user_files：用户视角文件，保存 user_id、parent_id、file_name、file_store_id
```

好处是：

- 同一份物理文件只在 MinIO 存一份。
- 多个用户或同一用户不同目录可以引用同一个 `file_store`。
- 秒传只需要新增 `user_files` 记录，不需要重新上传对象。
- 删除用户文件时只删除引用；永久删除时检查是否还有其他引用，没有引用才删除 `file_stores` 和 MinIO 对象。
- 用户空间按用户文件引用计费，创建引用时增加空间，永久删除时释放空间。

---

## 功能概览

| 模块 | 能力 |
| --- | --- |
| 用户 | 注册、登录、刷新 Token、获取用户信息、管理员更新会员/空间配置 |
| 文件上传 | 小文件普通上传、大文件 MinIO 分块直传、断点续传、分块完整性校验 |
| 秒传 | 抽样预筛、完整 hash 命中、随机片段 PoE、Redis 挑战会话、MinIO Range 校验 |
| 文件管理 | 目录列表、创建文件夹、批量删除、回收站、恢复、永久删除 |
| 下载 | 私有文件预签名下载、分享文件预签名下载 |
| 分享 | 创建文件分享、可选提取码、可选过期时间 |

---

## 核心接口

基础路径：

```text
/api/v1
```

### 公开接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/user/register` | 注册 |
| POST | `/user/login` | 登录，返回 Access Token 和 Refresh Token |
| POST | `/user/refresh` | 刷新 Access Token |
| POST | `/share/download` | 通过分享 token 和提取码获取下载地址 |

### 文件与上传接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/file/quick-check` | 抽样 hash 快速判断是否可能秒传 |
| POST | `/file/instant-upload/prepare` | 完整 hash 命中后生成随机片段挑战 |
| POST | `/file/instant-upload` | 提交随机片段证明，校验通过后秒传落库 |
| POST | `/file/upload` | 小文件普通上传 |
| POST | `/file/upload/init` | 初始化大文件分块上传 |
| POST | `/file/upload/sign-part` | 获取某个分块的 MinIO 预签名上传 URL |
| POST | `/file/upload/report-part` | 回报已上传分块 ETag |
| POST | `/file/upload/complete` | 完成 MinIO 分块合并并落库 |
| GET | `/file/download/:ID` | 获取私有文件下载 URL |

### 文件管理接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/file/list?parent_id=` | 查询当前目录文件和文件夹 |
| POST | `/file/upload-folder` | 创建文件夹 |
| DELETE | `/file/delete/:ID` | 删除文件或文件夹到回收站 |
| POST | `/file/batch-delete` | 批量删除 |
| GET | `/file/recycle` | 查看回收站 |
| POST | `/file/restore/:ID` | 从回收站恢复 |
| DELETE | `/file/permanent-delete/:ID` | 永久删除并释放空间 |
| POST | `/share/create` | 创建文件分享 |

---

## 项目结构

```text
cmd/                    启动入口
config/                 配置文件
internal/api/           HTTP handler
internal/app/           依赖组装和应用启动
internal/initialize/    MySQL、Redis、MinIO 初始化
internal/middleware/    JWT、超时、管理员中间件
internal/model/         GORM 模型
internal/oss/           MinIO 封装
internal/ratelimit/     Redis 令牌桶限流组件
internal/repository/    数据访问层
internal/router/        路由注册
internal/service/       核心业务逻辑
pkg/                    日志、响应、hash、JWT 等工具
frontend/               Vue 3 前端
```

---

## 运行说明

1. 修改配置：

```text
config/config.yaml
```

配置 MySQL、Redis、MinIO、JWT 等参数。

2. 启动后端：

```bash
go run ./cmd
```

3. 启动前端：

```bash
cd frontend
npm install
npm run dev
```
