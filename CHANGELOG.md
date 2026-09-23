# Changelog

## 1.1.0 — 2026-09-23

### Changed (server-side API)

- The anota API (production change of 2026-09-23) no longer returns a webhook's full signing
  secret when listing webhooks (`GET /api/v1/forms/{formId}/webhooks`, MCP `list_webhooks`).
  Each row of `webhooks` now carries `secretHint` (`whsec_…` plus the last 4 characters, or
  just `whsec_…` for short secrets) and `secretNote` (explains that the secret is shown once,
  at creation) in place of `secret`.
- `ListWebhooks` passes the server's JSON through unchanged, so code that read `secret` from a
  list row now gets no value. Read the secret from the `AddWebhook` response and store it.
- `AddWebhook` is unchanged: its response (`id`, `formId`, `url`, `secret`, `note`) is the only
  place the full secret ever appears. To replace a lost secret, delete the webhook and add it
  again.
- Method signatures are unchanged; the doc comments and READMEs now describe the new shape.

Released as a v1 minor on purpose: no SDK code or signature changed (`ListWebhooks` returns the
server JSON as-is), so moving the module path to `/v2` would force every caller to change imports
for a documentation-only update.

## 1.0.0

- Initial releases: 25-operation coverage of the anota API.
