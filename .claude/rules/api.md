---
paths:
  - "api/**"
  - "backend/internal/httpapi/**"
  - "web/src/lib/api/**"
---
# API rules

- `api/openapi.yaml` is the contract. Change it first, keep `docs/05-api.md` in sync when
  semantics change, run `make gen`, then implement. `redocly lint` must stay clean.
- Conventions (`docs/05-api.md` §1): `/api/v1`, camelCase JSON, enum values = database values,
  UUIDv7 ids, RFC 3339 UTC timestamps, keyset pagination (`limit`, `cursor`, `nextCursor`).
- Errors are Problem Details with one of the `ProblemCode` values and the documented extension
  members (`errors`, `conflicts`, `locked`, `blockers`, `reasons`). Map every error in one place.
- Every content mutation returns `MutationResult` and publishes the same rows as realtime
  events. Long work returns 202 with a resource to poll (sync run, export).
- Views (`getPfdView`, `getPfmeaView`, `getControlPlanView`) stream the JSON produced by one SQL
  statement through a custom strict-server response type; a contract test decodes them into
  the generated types with `DisallowUnknownFields`.
- Not-yet-implemented operations live in `httpapi/unimplemented.go` and return 501; move a
  method out when its milestone implements it.
- Mutations require `Content-Type: application/json`, a valid session cookie and an `Origin`
  equal to `APP_BASE_URL`; echo `X-Request-Id`.
