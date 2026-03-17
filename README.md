# rigel-console

Console application gateway and minimal frontend.

## Language

Go

## Current Stage

Slimmed down to the minimum UI/API shell for recommendation output.

## Implemented

- aggregates `rigel-build-engine` over HTTP
- exposes a price-catalog-first recommendation flow through `POST /catalog/recommend`
- serves an embedded frontend page for recommendation display

## Routes

- `GET /healthz`
- `POST /catalog/recommend`
- `GET /`

## Notes

- Console does not implement pricing, aggregation, compatibility, or AI logic.
- Console relies only on `rigel-build-engine` being reachable over HTTP.
- The homepage defaults to `UI params -> build-engine price catalog -> build-engine AI analysis result`.
