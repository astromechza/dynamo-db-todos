# Gemini Guidelines

This document provides instructions for Gemini on how to interact with and contribute to this project.

## Project Overview

This project is a simple Go HTML application that provides a to-do list over HTTP, backed by Amazon DynamoDB. It can be run locally as a standard web server or deployed as an AWS Lambda function.

## Key Technologies

*   **Go**: The programming language used for the application.
*   **HTML**: The language used for the web interface.
*   **Amazon DynamoDB**: The NoSQL database service used for data storage.
*   **AWS Lambda**: The serverless computing platform used for deployment.
*   **Amazon Bedrock**: The service used for generating todos.

## Project Structure

The project follows a simple structure:

*   `main.go`: The main application file.
*   `go.mod`, `go.sum`: Go module files for managing dependencies.
*   `Dockerfile`: For containerizing the application.
*   `.github/workflows/release.yml`: GitHub Actions workflow for releases.
*   `.goreleaser.yml`: Configuration for GoReleaser.

## How to Run

To run the application locally, follow these steps:

1.  **Set Environment Variables**:
    ```bash
    export AWS_REGION=your-region
    # export AWS_DYNAMODB_TABLE=your-table-name # Optional
    # export AWS_BEDROCK_MODEL_NAME=amazon.titan-text-lite-v1 # Optional
    ```
2.  **Run the Application**:
    ```bash
    go run main.go
    ```
3.  **Access the Application**:
    Open your web browser and navigate to `http://localhost:8080`.

## How to Deploy

To deploy the application to AWS Lambda, follow these steps:

1.  **Build the Executable**:
    ```bash
    GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
    ```
2.  **Create the Zip Package**:
    ```bash
    zip lambda.zip bootstrap
    ```
3.  **Deploy to Lambda**:
    *   Create a new Lambda function with a Go runtime.
    *   Upload the `lambda.zip` file as the function code.
    *   Set the handler to `bootstrap`.
    *   Configure the necessary environment variables (`AWS_REGION`, `AWS_DYNAMODB_TABLE`).
    *   Set up an API Gateway trigger to receive HTTP requests.

## Environment Variables

- `AWS_REGION`: The AWS region to use. This is required.
- `AWS_DYNAMODB_TABLE`: The name of the DynamoDB table to use. If not set, the application will start, but adding, deleting, and listing todos will be disabled.
- `AWS_BEDROCK_MODEL_NAME`: The name of the AWS Bedrock model to use for generating todos. If not set, the generate feature will be disabled. Example: `amazon.titan-text-lite-v1`.
- `MOTD`: An optional message of the day to display on the page.

## Versioning

This project uses semantic versioning (e.g., `v1.2.0`). Before creating a new release, please check the existing tags to determine the next appropriate version number.

## Code Quality

After making any changes to the Go code, please run the following commands to ensure code quality and correctness:

```bash
go build
go fmt
go vet
```
