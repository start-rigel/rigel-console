# rigel-console

`rigel-console` 是当前系统的最小前端与 API 入口。

## 当前职责

- 接收用户输入
- 调用 `rigel-build-engine`
- 展示推荐结果

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

## 当前目标

当前模块保持尽量薄：

`用户输入 -> build-engine -> 展示结果`

## 说明

当前 console 不承担抓取、聚合、AI 请求构建等核心逻辑。
这些逻辑统一由 `rigel-build-engine` 负责。
