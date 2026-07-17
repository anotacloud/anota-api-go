# anota-api-go · Cliente oficial de Go para la API de [anota](https://anota.cloud)

**[Read me in English](README.md)** · [Referencia interactiva de la API](https://anota.cloud/developers) · [Todos los SDK](https://github.com/anotacloud/anota-api)

![CI](https://github.com/anotacloud/anota-api-go/actions/workflows/ci.yml/badge.svg)

Crea y publica formularios, edita campos y lógica condicional, lee y escribe
respuestas y conecta webhooks: todo lo que la API REST de anota puede hacer,
desde Go.

Es un envoltorio delgado y sin dependencias: solo la biblioteca estándar de Go
(`net/http`, `encoding/json`). Cada llamada devuelve el JSON del servidor
decodificado en un `map[string]any`, así que nunca te bloquea un modelo de
respuesta que falte.

## Instalación

```sh
go get github.com/anotacloud/anota-api-go@v1.0.0
```

O descarga el [ZIP](https://github.com/anotacloud/anota-api-go/archive/refs/heads/main.zip)
/ [Tarball](https://github.com/anotacloud/anota-api-go/archive/refs/heads/main.tar.gz).

Luego impórtalo:

```go
import anota "github.com/anotacloud/anota-api-go"
```

## Inicio rápido

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

	form, err := client.CreateForm(ctx, "Formulario de contacto", []map[string]any{
		{"type": "text", "label": "Nombre", "required": true},
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

Hay un script completo y ejecutable en [`examples/endtoend`](examples/endtoend/main.go):

```sh
ANOTA_API_KEY=anota_sk_... go run ./examples/endtoend
```

## Autenticación

Crea una clave de API en tu workspace en https://anota.cloud/api-keys y pásala
al cliente. Las claves tienen el aspecto `anota_sk_…` y también habilitan el
conector MCP de Claude. Apunta a otro entorno estableciendo `client.BaseURL`
(por defecto `https://anota.cloud/api/v1`) y proporciona tu propio
`client.HTTPClient` para tiempos de espera o proxies personalizados.

## Todos los métodos

Cada método recibe un `context.Context` primero y devuelve `(map[string]any, error)`.

| # | Método | HTTP |
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

`fields`/`field` son `map[string]any` con las claves `type`, `label`,
`required?`, `options?`, `rows?`, `columns?`. `rules`/`rule` llevan `match`,
`if` y `then`. `answers` es un `map[string]any` indexado por id de campo, con
valores `string` o `[]string`. Para `ListSubmissions`, un `page`/`pageSize` no
positivo toma el valor por defecto 1/25 y un `status` vacío lista todos los
estados; para `ListTemplates` un `language` vacío toma el valor por defecto
`"es"`.

## Errores

Las respuestas que no son 2xx devuelven un `*APIError` que lleva el código de
estado HTTP y el mensaje del servidor (extraído del `detail` de problem-details,
luego `title`, luego el cuerpo sin procesar). Los fallos de red se manifiestan
como el error nativo de `net/http`, no como un `*APIError`.

```go
form, err := client.GetForm(ctx, "no-existe")
if err != nil {
	var apiErr *anota.APIError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.Status, apiErr.Message)
	}
}
```

Nota: una vez que un formulario se ha publicado, sus campos existentes quedan
bloqueados (`EditField`/`DeleteField` devuelven 400); siempre puedes usar
`AddFields`.

## Licencia

MIT — consulta [LICENSE](LICENSE).
