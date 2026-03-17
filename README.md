# rigel-console

Console application gateway, minimal frontend, and orchestration layer.

## Language

Go

## Current Stage

Phase 6 minimum viable implementation.

## Implemented

- aggregates build-engine over HTTP
- exposes a price-catalog-first recommendation flow through `POST /catalog/recommend`
- exposes `POST /build/generate`, `GET /build/{id}`, and `GET /parts/search`
- serves an embedded frontend page for price-catalog recommendation
- serves admin pages backed by collector/build-engine APIs for products, parts, catalog, and jobs
- supports triggering JD collection and retrying collector jobs from the admin surface
- supports triggering the `mvp_base` batch collection preset from the admin surface
- surfaces batch collection skip/abort counts so admin users can see when JD risk control cut a run short
- exposes an admin price-catalog view backed by `rigel-build-engine /api/v1/catalog/prices`
- returns a slim user-facing build payload centered on selected models, price, warnings, alternatives, and AI advice

## Routes

- `GET /healthz`
- `POST /catalog/recommend`
- `POST /build/generate`
- `GET /build/{id}`
- `GET /parts/search?keyword=ryzen&limit=10`
- `POST /api/admin/collect/search`
- `POST /api/admin/collect/batch`
- `GET /api/admin/catalog/prices?use_case=gaming&build_mode=mixed`
- `GET /api/admin/products?keyword=4060&limit=10`
- `GET /api/admin/parts?keyword=ryzen&limit=10`
- `GET /api/admin/jobs?limit=10`
- `POST /api/admin/jobs/{id}/retry`
- `GET /`
- `GET /admin/products`
- `GET /admin/parts`
- `GET /admin/catalog`
- `GET /admin/jobs`

## Notes

- Console does not implement compatibility logic.
- Console relies on `rigel-build-engine` and `rigel-jd-collector` being reachable over HTTP.
- The homepage now defaults to `price catalog -> AI recommendation draft`; the older `/build/generate` route is kept for the structured build flow.
- Admin pages are intentionally lightweight and proxy existing service APIs instead of reading databases directly.
- Admin product management is now expected to default to real JD data and can narrow further to JD self-operated products.
