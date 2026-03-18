# rigel-console

`rigel-console` 是当前系统的最小前端与 API 入口。

当前对外站点域名规划：

- `givezj8.cn`

## 当前职责

- 接收用户输入
- 调用 `rigel-build-engine`
- 展示推荐结果
- 提供中英文切换页面体验

## 不负责什么

- 不直接抓取外部平台数据
- 不直接做价格清单整理
- 不直接做 AI 分析

## 当前输入

来自用户界面的参数，例如：

- 预算
- 用途
- 品牌偏好
- 特殊要求
- 补充说明

## 当前输出

面向用户的推荐结果页面或推荐结果接口响应。

## 当前接口

- `GET /healthz`
- `POST /catalog/recommend`
- `GET /`

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

### 2. 推荐请求

请求：

```bash
curl -X POST http://localhost:18084/catalog/recommend \
  -H "Content-Type: application/json" \
  -d '{
    "budget": 6000,
    "use_case": "gaming",
    "build_mode": "mixed",
    "brand_preference": {
      "cpu": "amd",
      "gpu": "nvidia"
    },
    "special_requirements": [
      "wifi_motherboard"
    ],
    "notes": "1080p 游戏为主，希望整机尽量均衡"
  }'
```

响应示例：

```json
{
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
- 当前返回结构以 `catalog_item_count`、`selection`、`advice` 为主，不是旧版 `parts` / `total_price` 平铺结构

### 3. 页面入口

```bash
curl -I http://localhost:18084/
curl -I http://localhost:18084/keywords
curl -I http://localhost:18084/keywords/new
curl -I http://localhost:18084/keywords/import
```

## 当前页面

- `GET /`
- `GET /keywords`
- `GET /keywords/new`
- `GET /keywords/{id}/edit`
- `GET /keywords/import`

## 当前页面要求

- 页面文案支持中文 / English 切换
- 语言切换优先使用前端本地状态持久化
- 不要求额外的后端国际化接口

## 当前目标

当前模块保持尽量薄：

`用户输入 -> build-engine -> 展示结果`

## 说明

当前 console 不承担抓取、聚合、AI 请求构建等核心逻辑。
这些逻辑统一由 `rigel-build-engine` 负责。
