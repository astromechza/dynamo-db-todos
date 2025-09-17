package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type Todo struct {
	Id                 string
	Text               string
	CreatedAtEpoch     int64
	CreatedAtFormatted string `dynamodbav:"-"`
}

var (
	dynamoClient      *dynamodb.Client
	bedrockClient     *bedrockruntime.Client
	tableName         string
	bedrockModelName  string
	messageOfTheDay   string
)

const indexTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Todo List</title>
    <style>
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            border: 1px solid #dddddd;
            text-align: left;
            padding: 8px;
        }
        th {
            background-color: #f2f2f2;
        }
        .actions {
            width: 1%;
            white-space: nowrap;
        }
    </style>
</head>
<body>
    <h1>Todo List</h1>
    {{if .MOTD}}
    <p><b>Message of the day:</b> {{.MOTD}}</p>
    {{end}}
    <form action="/add" method="post" style="display:inline-block; margin-bottom: 20px;">
        <input type="text" name="text" size="50">
        <input type="submit" value="Add">
    </form>
    <form action="/generate" method="post" style="display:inline-block;">
        <input type="submit" value="Generate">
    </form>
    <table>
        <thead>
            <tr>
                <th>Timestamp</th>
                <th>Todo</th>
                <th class="actions">Actions</th>
            </tr>
        </thead>
        <tbody>
            {{range .}}
            <tr>
                <td>{{.CreatedAtFormatted}}</td>
                <td>{{.Text}}</td>
                <td class="actions">
                    <form action="/delete" method="post" style="display:inline;">
                        <input type="hidden" name="id" value="{{.Id}}">
                        <input type="hidden" name="createdAtEpoch" value="{{.CreatedAtEpoch}}">
                        <input type="submit" value="Done">
                    </form>
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
</body>
</html>
`

// TitanTextRequest is the request payload for the Amazon Titan Text model.
type TitanTextRequest struct {
	InputText            string `json:"inputText"`
	TextGenerationConfig struct {
		MaxTokenCount int      `json:"maxTokenCount"`
		StopSequences []string `json:"stopSequences"`
		Temperature   float64  `json:"temperature"`
		TopP          float64  `json:"topP"`
	} `json:"textGenerationConfig"`
}

// TitanTextResponse is the response payload from the Amazon Titan Text model.
type TitanTextResponse struct {
	InputTextTokenCount int `json:"inputTextTokenCount"`
	Results             []struct {
		TokenCount       int    `json:"tokenCount"`
		OutputText       string `json:"outputText"`
		CompletionReason string `json:"completionReason"`
	} `json:"results"`
}

func main() {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		log.Fatal("AWS_REGION environment variable not set")
	}
	tableName = os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		log.Fatal("DYNAMODB_TABLE environment variable not set")
	}
	bedrockModelName = os.Getenv("AWS_BEDROCK_MODEL_NAME")
	if bedrockModelName == "" {
		bedrockModelName = "amazon.titan-text-lite-v1"
	}
	messageOfTheDay = os.Getenv("MOTD")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	bedrockClient = bedrockruntime.NewFromConfig(cfg)

	http.HandleFunc("/", listHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/generate", generateHandler)

	log.Printf("Server starting on port 8080...")
	if messageOfTheDay != "" {
		log.Printf("Message of the day: %s", messageOfTheDay)
	}
	log.Printf("Using bedrock model: %s", bedrockModelName)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	todos, err := getTodos()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get todos: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, struct {
		Todos []Todo
		MOTD  string
	}{
		Todos: todos,
		MOTD:  messageOfTheDay,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	text := r.FormValue("text")
	if text == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Security hardening: Validate the input text.
	if utf8.RuneCountInString(text) > 200 {
		http.Error(w, "Todo text must be 200 characters or less.", http.StatusBadRequest)
		return
	}
	for _, r := range text {
		if !unicode.IsPrint(r) {
			http.Error(w, "Todo text contains non-printable characters.", http.StatusBadRequest)
			return
		}
	}

	err := addTodo(text)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to add todo: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	prompt := "Generate a short, fake to-do list item. It must be a single sentence, not a list."

	request := TitanTextRequest{
		InputText: prompt,
	}
	request.TextGenerationConfig.MaxTokenCount = 50
	request.TextGenerationConfig.StopSequences = []string{}
	request.TextGenerationConfig.Temperature = 0.9
	request.TextGenerationConfig.TopP = 1.0

	payload, err := json.Marshal(request)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal request: %v", err), http.StatusInternalServerError)
		return
	}

	modelID := bedrockModelName
	output, err := bedrockClient.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
		Body:        payload,
		ModelId:     aws.String(modelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to invoke bedrock model: %v", err), http.StatusInternalServerError)
		return
	}

	var response TitanTextResponse
	err = json.Unmarshal(output.Body, &response)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to unmarshal response: %v", err), http.StatusInternalServerError)
		return
	}

	if len(response.Results) == 0 {
		http.Error(w, "empty response from bedrock", http.StatusInternalServerError)
		return
	}

	generatedText := response.Results[0].OutputText

	// The model might return the text with leading/trailing whitespace or quotes. Clean it up.
	generatedText = strings.TrimSpace(generatedText)
	generatedText = strings.Trim(generatedText, "\"")

	err = addTodo(generatedText)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to add generated todo: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")
	createdAtEpochStr := r.FormValue("createdAtEpoch")

	createdAtEpoch, err := strconv.ParseInt(createdAtEpochStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid createdAtEpoch", http.StatusBadRequest)
		return
	}

	err = deleteTodo(id, createdAtEpoch)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete todo: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getTodos() ([]Todo, error) {
	out, err := dynamoClient.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, err
	}

	var todos []Todo
	err = attributevalue.UnmarshalListOfMaps(out.Items, &todos)
	if err != nil {
		return nil, err
	}

	// Format the timestamp for display and sort the todos by creation time (newest first).
	for i := range todos {
		todos[i].CreatedAtFormatted = time.Unix(todos[i].CreatedAtEpoch, 0).Format("2006-01-02 15:04:05")
	}
	sort.Slice(todos, func(i, j int) bool {
		return todos[i].CreatedAtEpoch > todos[j].CreatedAtEpoch
	})

	return todos, nil
}

func addTodo(text string) error {
	todo := Todo{
		Id:             uuid.New().String(),
		Text:           text,
		CreatedAtEpoch: time.Now().Unix(),
	}

	item, err := attributevalue.MarshalMap(todo)
	if err != nil {
		return err
	}

	_, err = dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	return err
}

func deleteTodo(id string, createdAtEpoch int64) error {
	_, err := dynamoClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"Id":             &types.AttributeValueMemberS{Value: id},
			"CreatedAtEpoch": &types.AttributeValueMemberN{Value: strconv.FormatInt(createdAtEpoch, 10)},
		},
	})
	return err
}
