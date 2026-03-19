# AGENTS.md

## Core Rule

This repository must follow the shared project constraints defined in:

- `../rigel-core/AGENTS.md`

## Core Docs Location

Overall project documentation, workspace-level architecture, database notes, and deployment files are centralized in:

- `../rigel-core`

## Usage Rule

When working in this repository:

1. Read and follow `../rigel-core/AGENTS.md` first.
2. Treat `rigel-core` as the source of truth for workspace-level documentation.
3. Use this repository's local README and code layout only as module-specific supplements.
4. If a local module document conflicts with `rigel-core`, pause and reconcile instead of guessing.

## Security Supplement

1. `rigel-console` 当前是默认公网入口，新增公开接口时必须先考虑匿名配额、缓存、冷却、挑战和日志审计。
2. 后台 `/admin` 与 `/admin/api/*` 默认按私网 / VPN 访问设计；不要为了方便开发或调试直接移除来源限制。
3. 所有发往 `rigel-build-engine` 的高成本内部请求，默认继续透传服务 token，不要绕过内部鉴权。
4. 前台匿名链路默认继续携带 `device_fingerprint`，不要退回到只依赖 Cookie 的风控方案。
