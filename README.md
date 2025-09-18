# DynamoDB Todo List

A simple Go HTML application that provides a to-do list over HTTP, backed by Amazon DynamoDB.

## Prerequisites

- Go installed
- AWS CLI configured with appropriate credentials. Your AWS identity will need `dynamodb:Scan`, `dynamodb:PutItem`, `dynamodb:DeleteItem` permissions for the table, and `bedrock:InvokeModel` permission for the `amazon.titan-text-lite-v1` model.

## Setup

1.  **Enable Model Access in Amazon Bedrock:**
    - Navigate to the Amazon Bedrock console in your AWS Region.
    - In the bottom-left navigation pane, click on **Model access**.
    - Click **Manage model access** and enable access for **Titan Text Lite** (`amazon.titan-text-lite-v1`).

2.  **Create the DynamoDB Table:**

    Run the following command to create the necessary DynamoDB table. Replace `your-table-name` and `your-region` with your desired values.

    ```bash
    aws dynamodb create-table \
        --table-name Todos \
        --attribute-definitions \
            AttributeName=Id,AttributeType=S \
            AttributeName=CreatedAtEpoch,AttributeType=N \
        --key-schema \
            AttributeName=Id,KeyType=HASH \
            AttributeName=CreatedAtEpoch,KeyType=RANGE \
        --billing-mode PAY_PER_REQUEST \
        --region eu-central-1
    ```

## Deployment

The application can be run locally as a standard web server or deployed as an AWS Lambda function.

### Local

1.  **Set Environment Variables:**

    The environment variables are optional and enable specific features.

    ```bash
    # export AWS_REGION=your-region # Optional
    # export AWS_DYNAMODB_TABLE=your-table-name # Optional
    # export AWS_BEDROCK_MODEL_NAME=amazon.titan-text-lite-v1 # Optional
    ```

2.  **Run the Application:**

    Start the web server with the following command:

    ```bash
    go run main.go
    ```

3.  **Access the Application:**

    Open your web browser and navigate to `http://localhost:8080`.

### AWS Lambda

To deploy the application to AWS Lambda, you need to build a Linux binary and package it as a zip file.

1.  **Build the Executable:**

    Compile the application for the Lambda environment:

    ```bash
    GOOS=linux GOARCH=amd64 go build -o bootstrap main.go
    ```

2.  **Create the Zip Package:**

    Package the executable into a zip file:

    ```bash
    zip lambda.zip bootstrap
    ```

3.  **Deploy to Lambda:**

    - Create a new Lambda function with a Go runtime.
    - Upload the `lambda.zip` file as the function code.
    - Set the handler to `bootstrap`.
    - Configure the necessary environment variables (`AWS_REGION`, `AWS_DYNAMODB_TABLE`).
    - Set up an API Gateway trigger to receive HTTP requests.

## Environment Variables

- `AWS_REGION`: The AWS region to use. If not set, the application will attempt to determine the region from the environment (e.g., IAM role).
- `AWS_DYNAMODB_TABLE`: The name of the DynamoDB table to use. If not set, the application will start, but adding, deleting, and listing todos will be disabled.
- `AWS_BEDROCK_MODEL_NAME`: The name of the AWS Bedrock model to use for generating todos. If not set, the generate feature will be disabled. Example: `amazon.titan-text-lite-v1`.
- `MOTD`: An optional message of the day to display on the page.
