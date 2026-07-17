# anota-api-go · Official Go client for the [anota](https://anota.cloud) API

**[Léeme en español](README.es.md)** · [Interactive API reference](https://anota.cloud/developers) · [All SDKs](https://github.com/anotacloud/anota-api)

![CI](https://github.com/anotacloud/anota-api-go/actions/workflows/ci.yml/badge.svg)

Create and publish forms, edit fields and conditional logic, read and write
submissions, and wire webhooks — everything the anota REST API can do, from Go.

It is a thin, dependency-free wrapper: only the Go standard library
(`net/http`, `encoding/json`). Every call returns the server's JSON decoded into
a `map[string]any`, so you are never blocked by a missing response model.

## Install

```sh
go get github.com/anotacloud/anota-api-go@v1.0.0
```

Or download the [ZIP](https://github.com/anotacloud/anota-api-go/archive/refs/heads/main.zip)
/ [Tarball](https://github.com/anotacloud/anota-api-go/archive/refs/heads/main.tar.gz).

Then import it:

```go
import anota "github.com/anotacloud/anota-api-go"
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	anota "github.com/anotacloud/anota-api-go"
)

func main() {
	client := anota.New(os.Getenv("ANOTA_API_KEY"))
	ctx := context.Background()

	form, err := client.CreateForm(ctx, "Contact form", []map[string]any{
		{"type": "text", "label": "Name", "required": true},
	}, "")
	if err != nil {
		log.Fatal(err)
	}
	formID := form["id"].(string)

	if _, err := client.PublishForm(ctx, formID); err != nil {
		log.Fatal(err)
	}

	subs, err := client.ListSubmissions(ctx, formID, 1, 25, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(subs)
}
```

A full runnable script lives in [`examples/endtoend`](examples/endtoend/main.go):

```sh
ANOTA_API_KEY=anota_sk_... go run ./examples/endtoend
```

## Authentication

Create an API key in your workspace at https://anota.cloud/api-keys and pass it
to the client. Keys look like `anota_sk_…` and also power the Claude MCP
connector. Point at a different environment by setting `client.BaseURL` (default
`https://anota.cloud/api/v1`), and supply your own `client.HTTPClient` for
custom timeouts or proxies.

## All methods

Every method takes a `context.Context` first and returns `(map[string]any, error)`.

| # | Method | HTTP |
|---|---|---|
| 1 | `ListForms(ctx)` | `GET /forms` |
| 2 | `CreateForm(ctx, title, fields, description)` | `POST /forms` |
| 3 | `GetForm(ctx, formID)` | `GET /forms/{formId}` |
| 4 | `AddFields(ctx, formID, fields)` | `POST /forms/{formId}/fields` |
| 5 | `EditField(ctx, formID, fieldID, field)` | `PATCH /forms/{formId}/fields/{fieldId}` |
| 6 | `DeleteField(ctx, formID, fieldID)` | `DELETE /forms/{formId}/fields/{fieldId}` |
| 7 | `PublishForm(ctx, formID)` | `POST /forms/{formId}/publish` |
| 8 | `RenameForm(ctx, formID, title)` | `PATCH /forms/{formId}` |
| 9 | `SetPdfTemplate(ctx, formID, key)` | `PUT /forms/{formId}/pdf-template` |
| 10 | `DeleteForm(ctx, formID)` | `DELETE /forms/{formId}` |
| 11 | `CloneForm(ctx, formID)` | `POST /forms/{formId}/clone` |
| 12 | `AddLogicRules(ctx, formID, rules)` | `POST /forms/{formId}/logic-rules` |
| 13 | `EditLogicRule(ctx, formID, ruleID, rule)` | `PUT /forms/{formId}/logic-rules/{ruleId}` |
| 14 | `DeleteLogicRule(ctx, formID, ruleID)` | `DELETE /forms/{formId}/logic-rules/{ruleId}` |
| 15 | `ListSubmissions(ctx, formID, page, pageSize, status)` | `GET /forms/{formId}/submissions` |
| 16 | `GetSubmission(ctx, submissionID)` | `GET /submissions/{submissionId}` |
| 17 | `CreateSubmission(ctx, formID, answers)` | `POST /forms/{formId}/submissions` |
| 18 | `SetSubmissionStatus(ctx, submissionID, status)` | `PATCH /submissions/{submissionId}/status` |
| 19 | `DeleteSubmission(ctx, submissionID)` | `DELETE /submissions/{submissionId}` |
| 20 | `SubmissionStats(ctx, formID)` | `GET /forms/{formId}/stats` |
| 21 | `ListTemplates(ctx, language)` | `GET /templates?language=` |
| 22 | `CreateFormFromTemplate(ctx, templateID)` | `POST /forms/from-template/{templateId}` |
| 23 | `ListWebhooks(ctx, formID)` | `GET /forms/{formId}/webhooks` |
| 24 | `AddWebhook(ctx, formID, url)` | `POST /forms/{formId}/webhooks` |
| 25 | `DeleteWebhook(ctx, formID, webhookID)` | `DELETE /forms/{formId}/webhooks/{webhookId}` |

`fields`/`field` are `map[string]any` with keys `type`, `label`, `required?`,
`options?`, `rows?`, `columns?`. `rules`/`rule` carry `match`, `if`, and `then`.
`answers` is a `map[string]any` keyed by field id, with `string` or `[]string`
values. For `ListSubmissions`, a non-positive `page`/`pageSize` defaults to 1/25
and an empty `status` lists all statuses; for `ListTemplates` an empty
`language` defaults to `"es"`.

## Errors

Non-2xx responses return an `*APIError` carrying the HTTP status and the
server's message (extracted from the problem-details `detail`, then `title`,
then the raw body). Network failures surface as the native `net/http` error,
not an `*APIError`.

```go
form, err := client.GetForm(ctx, "does-not-exist")
if err != nil {
	var apiErr *anota.APIError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.Status, apiErr.Message)
	}
}
```

Note: once a form has been published, its existing fields are locked
(`EditField`/`DeleteField` return 400); you can always `AddFields`.

## License

MIT — see [LICENSE](LICENSE).
