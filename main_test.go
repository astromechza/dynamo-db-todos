package main

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestListHandlerTemplateRendering(t *testing.T) {
	tmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	data := struct {
		Todos                []Todo
		MOTD                 string
		IsDynamoDBConfigured bool
		IsBedrockConfigured  bool
		Prefix               string
	}{
		Todos: []Todo{
			{Id: "1", Text: "Test Todo", CreatedAtEpoch: 1678886400},
		},
		MOTD:                 "Test MOTD",
		IsDynamoDBConfigured: true,
		IsBedrockConfigured:  true,
		Prefix:               "/test/",
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	output := buf.String()

	expectedActions := []string{
		`action="/test/add"`,
		`action="/test/generate"`,
		`action="/test/delete"`,
	}

	for _, expected := range expectedActions {
		if !strings.Contains(output, expected) {
			t.Errorf("expected to find %q in the template output, but did not", expected)
		}
	}
}
