// Command endtoend walks the full happy path of the anota API from Go:
// create a form, add a field, publish, submit an answer, and list submissions.
//
// Run it with your API key in the environment:
//
//	ANOTA_API_KEY=anota_sk_... go run ./examples/endtoend
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	anota "github.com/anotacloud/anota-api-go"
)

func main() {
	apiKey := os.Getenv("ANOTA_API_KEY")
	if apiKey == "" {
		log.Fatal("set ANOTA_API_KEY (create a key at https://anota.cloud/api-keys)")
	}

	client := anota.New(apiKey)
	ctx := context.Background()

	form, err := client.CreateForm(ctx, "Contact form", []map[string]any{
		{"type": "text", "label": "Name", "required": true},
	}, "Created by the anota Go SDK example")
	if err != nil {
		log.Fatalf("createForm: %v", err)
	}
	formID, _ := form["id"].(string)
	fmt.Printf("Created form %s\n", formID)

	if _, err := client.AddFields(ctx, formID, []map[string]any{
		{"type": "email", "label": "Email", "required": true},
	}); err != nil {
		log.Fatalf("addFields: %v", err)
	}
	fmt.Println("Added an email field")

	if _, err := client.PublishForm(ctx, formID); err != nil {
		log.Fatalf("publishForm: %v", err)
	}
	fmt.Println("Published the form")

	submission, err := client.CreateSubmission(ctx, formID, map[string]any{
		"f_1": "Ada Lovelace",
		"f_2": "ada@example.com",
	})
	if err != nil {
		log.Fatalf("createSubmission: %v", err)
	}
	fmt.Printf("Created submission %v\n", submission["id"])

	submissions, err := client.ListSubmissions(ctx, formID, 1, 25, "")
	if err != nil {
		log.Fatalf("listSubmissions: %v", err)
	}
	fmt.Printf("Submissions so far: %v\n", submissions)
}
