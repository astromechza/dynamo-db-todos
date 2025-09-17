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

2.  **Set Environment Variables:**

    Export the following environment variables, replacing the values with the ones you used in the previous step:

    ```bash
    export AWS_REGION=your-region
    export DYNAMODB_TABLE=your-table-name
    ```

3.  **Run the Application:**

    Start the web server with the following command:

    ```bash
    go run main.go
    ```

4.  **Access the Application:**

    Open your web browser and navigate to `http://localhost:8080`.

## Environment Variables

- `AWS_REGION`: The AWS region to use.
- `DYNAMODB_TABLE`: The name of the DynamoDB table to use.
- `AWS_BEDROCK_MODEL_NAME`: The name of the AWS Bedrock model to use for generating todos. Defaults to `amazon.titan-text-lite-v1`.
- `MOTD`: An optional message of the day to display on the page.
