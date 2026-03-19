# rigel-console

`rigel-console` 是当前系统的前台用户页、后台管理页和推荐 API 入口。

当前对外产品名称：

- `给我装机吧`

当前对外站点域名规划：

- `givezj8.cn`

## 当前职责

- 提供匿名可直接使用的前台推荐页
- 提供必须登录后访问的后台管理页
- 维护 `frontend/` 下的 React + Vite 前端工程，并将构建产物内嵌进 Go 服务
- 管理匿名会话和后台登录态
- 接收用户输入
- 调用 `rigel-build-engine`
- 展示推荐结果
- 提供词库管理、导入导出和启停操作
- 提供 JD 定时采集配置页面和后台代理接口

## 不负责什么

- 不直接抓取外部平台数据
- 不直接做价格清单整理
- 不直接做 AI 分析

## 当前输入

来自用户界面的参数，例如：

- 预算
- 用途
- 补充说明

当前首页默认不再让用户显式选择品牌偏好和装机模式：

- `build_mode` 固定使用默认值
- `brand_preference` 和 `special_requirements` 由后端接受，但不作为首页第一版显式输入项

## 当前输出

面向用户的推荐结果页面或推荐结果接口响应。

前台推荐结果页当前重点展示：

- 估算总价
- 已选配件清单
- 当前价格样本数量
- 推荐理由
- 风险提示
- 后续升级建议

## 当前接口

- `GET /healthz`
- `GET /api/v1/bootstrap`
- `GET /api/v1/session/anonymous`
- `POST /api/v1/challenge/verify`
- `POST /catalog/recommend`
- `GET /`
- `GET /admin/login`
- `POST /admin/login`
- `POST /admin/logout`
- `GET /admin`
- `GET /admin/jd-schedule`
- `GET /admin/api/v1/jd/schedule`
- `PUT /admin/api/v1/jd/schedule`
- `GET /admin/api/v1/keyword-seeds`
- `POST /admin/api/v1/keyword-seeds`
- `PUT /admin/api/v1/keyword-seeds/{id}`
- `POST /admin/api/v1/keyword-seeds/{id}/enable`
- `POST /admin/api/v1/keyword-seeds/{id}/disable`
- `POST /admin/api/v1/keyword-seeds/import`
- `GET /admin/api/v1/keyword-seeds/template`
- `GET /admin/api/v1/keyword-seeds/export`

## 配置方式

当前服务默认读取：

```text
configs/config.yaml
```

前端源码位于：

```text
frontend/
```

前端生产构建产物默认输出到：

```text
internal/app/web/dist
```

当前页面主路径说明：

- `frontend/` 是唯一页面源码目录
- `internal/app/web/dist` 是唯一运行中的前端产物目录
- 旧的 `internal/app/web/*.html` 静态页面已移除，不再作为维护路径

启动示例：

```bash
go run ./cmd/server -config ./configs/config.yaml
```

默认开发配置：

- 匿名会话小时额度：`5`
- 单 IP 小时额度：`20`
- 单设备指纹小时额度：`12`
- 匿名冷却秒数：`60`
- 挑战通过放行秒数：`900`

当前关键配置补充：

- `postgres_dsn` 必填，词库管理直接读写 PostgreSQL 中的 `rigel_keyword_seeds`
- `build_engine_token` 必填，`console -> build-engine` 默认强制走内部 token
- 后台登录不再允许默认弱口令；至少要提供 `admin_password` 或 `admin_password_hash`
- 推荐挑战如使用 Turnstile，需要同时配置 `challenge_provider=turnstile`、`challenge_site_key`、`challenge_secret`

当前安全规则：

- 前台推荐请求默认会附带设备指纹头 `X-Device-Fingerprint`
- 前台会先读取 `/api/v1/bootstrap`，按后端下发的挑战配置决定是否渲染挑战组件
- `rigel-console -> rigel-build-engine` 默认通过 `build_engine_token` 走内部 token 鉴权
- 后台页面与后台 API 默认只允许私网 / VPN 来源访问
- 后台登录态改为服务端 session，不再使用固定值 cookie
- 后台修改类接口默认要求 `X-CSRF-Token`
- 词库 CRUD / 导入导出直接使用 PostgreSQL 真源，不再走进程内存

## 开发约束

- `rigel-console` 默认是唯一公网入口；新增公开接口时，要先补匿名风控、冷却、缓存或挑战策略
- 不要为了联调方便移除 `/admin` 的私网 / VPN 限制；如果需要改公网策略，必须先同步 `rigel-core`
- 不要绕过 `build_engine_token` 直接请求 `rigel-build-engine` 高成本接口
- 不要把真实挑战密钥、内部 token、后台真实口令写进仓库配置或 README 示例

## 前端开发与构建

安装前端依赖：

```bash
cd frontend
npm install
```

本地开发前端：

```bash
cd frontend
npm run dev
```

构建并刷新 Go 内嵌静态资源：

```bash
cd frontend
npm run build
```

说明：

- React 前端负责渲染 `/`、`/admin/login`、`/admin`、`/admin/jd-schedule`、`/admin/keywords`、`/admin/keywords/new`、`/admin/keywords/{id}/edit`、`/admin/keywords/import`
- Go 仍然负责所有业务 API、Cookie、后台鉴权与静态资源分发
- `npm run build` 后必须重新启动 `rigel-console`，新的内嵌页面产物才会生效
- 前台与后台页面都支持本地日间 / 夜间主题切换，主题状态保存在浏览器本地

## 接口示例

### 1. 健康检查

请求：

```bash
curl http://localhost:18084/healthz
```

响应示例：

```json
{
  "status": "ok",
  "service": "rigel-console"
}
```

### 2. 匿名会话

请求：

```bash
curl http://localhost:18084/api/v1/session/anonymous
```

响应示例：

```json
{
  "anonymous_id": "anon_xxxxx",
  "cooldown_seconds": 0,
  "remaining_ai_requests": 5,
  "challenge_required": false,
  "challenge_passed": false,
  "risk_level": "normal",
  "session_expires_in_seconds": 2592000
}
```

### 3. 推荐请求

请求：

```bash
curl http://localhost:18084/api/v1/session/anonymous
curl -X POST http://localhost:18084/catalog/recommend \
  -H "Content-Type: application/json" \
  -H "X-Anonymous-Id: anon_xxxxx" \
  -H "X-Device-Fingerprint: fp_xxxxx" \
  -d '{
    "budget": 6000,
    "use_case": "gaming",
    "build_mode": "mixed",
    "notes": "1080p 游戏为主，希望整机尽量均衡"
  }'
```

响应示例：

```json
{
  "request_status": {
    "cache_hit": false,
    "remaining_ai_requests": 4,
    "cooldown_seconds": 0,
    "challenge_required": false,
    "risk_level": "normal"
  },
  "catalog_item_count": 24,
  "catalog_warnings": [],
  "selection": {
    "budget": 6000,
    "use_case": "gaming",
    "build_mode": "mixed",
    "estimated_total": 4206,
    "warnings": [
      "当前价格目录缺少这些类别的数据：MB、PSU、CASE、COOLER。"
    ],
    "selected_items": [
      {
        "category": "CPU",
        "display_name": "AMD 7500f",
        "normalized_key": "cpu-7500f",
        "sample_count": 3,
        "selected_price": 899,
        "median_price": 899,
        "source_platforms": ["jd"],
        "reasons": [
          "当前类别按 1200 元目标预算挑选了更接近中位价的型号。",
          "已参考 3 个价格样本。"
        ]
      },
      {
        "category": "GPU",
        "display_name": "NVIDIA rtx 4060",
        "normalized_key": "gpu-rtx-4060",
        "sample_count": 4,
        "selected_price": 2399,
        "median_price": 2399,
        "source_platforms": ["jd"],
        "reasons": [
          "当前类别按 3000 元目标预算挑选了更接近中位价的型号。",
          "已参考 4 个价格样本。"
        ]
      }
    ]
  },
  "advice": {
    "summary": "基于当前价格目录，这份 gaming 采购草案总价约 4206 元，核心组合为 AMD 7500f 和 NVIDIA rtx 4060。",
    "reasons": [
      "本次按 6000 元预算和 gaming 用途，从当前价格目录中挑选了更接近预算中心的型号。",
      "草案总价约 4206 元，优先参考了各型号的中位价和样本量。"
    ],
    "fit_for": [
      "1080p/2K 主流游戏场景"
    ],
    "risks": [
      "价格目录会随平台活动和库存变化波动，建议下单前重新抓取一次最新价格。"
    ],
    "upgrade_advice": [
      "如果游戏库会持续变大，优先把 SSD 升到 2TB。"
    ],
    "alternative_note": "如果你更看重品牌、静音或不同采购偏好，可以在同一份价格目录上再生成一版草案。"
  }
}
```

说明：

- console 当前会先向 build-engine 获取价格目录，再请求推荐草案
- 当前返回结构以 `request_status`、`catalog_item_count`、`selection`、`advice` 为主
- 相同参数优先返回缓存结果，不重复消耗匿名 AI 次数
- 高风险或缺失指纹的请求会返回 `challenge_required=true`
- 当前挑战验证接口为 `POST /api/v1/challenge/verify`

### 4. 后台登录

请求：

```bash
curl -X POST http://localhost:18084/admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }'
```

### 5. 页面入口

```bash
curl -I http://localhost:18084/
curl -I http://localhost:18084/admin/login
curl -I http://localhost:18084/admin
curl -I http://localhost:18084/admin/keywords
curl -I http://localhost:18084/admin/keywords/new
curl -I http://localhost:18084/admin/keywords/import
```

## 当前页面

- `GET /`
- `GET /admin/login`
- `GET /admin`
- `GET /admin/keywords`
- `GET /admin/keywords/new`
- `GET /admin/keywords/{id}/edit`
- `GET /admin/keywords/import`

## 当前页面要求

- 前台推荐页允许匿名直接访问
- 后台管理页必须先登录
- 后台管理页默认只允许私网 / VPN 访问
- 前台和后台页面路由明确分离
- 页面前端统一由 React 渲染，Go 侧只负责返回嵌入式 SPA 页面壳和 API

## 当前目标

当前模块当前已形成两条入口：

- 前台：`匿名用户输入 -> build-engine -> 展示结果`
- 后台：`登录 -> 词库管理 -> 导入导出 / 启停`

## 说明

当前 console 里的词库管理是第一版最小闭环实现：

- 后台登录态使用 Cookie
- 词库数据当前先在进程内存中维护
- Excel 模板下载、导出和导入已经可用

真正的抓取、聚合和 AI 分析逻辑仍由其他服务负责。
