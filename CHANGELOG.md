# Changelog

## Unreleased (next major)

### Breaking

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

Note: publishing this as v2.0.0 requires changing the module path in `go.mod` to `github.com/anotacloud/anota-api-go/v2` (Go semantic import versioning).

## 1.0.0

- Initial releases: 25-operation coverage of the anota API.
