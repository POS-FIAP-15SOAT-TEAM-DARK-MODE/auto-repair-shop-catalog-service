#!/bin/sh
# Creates the messaging resources of the catalog service in MiniStack (idempotent).
# Physical names: auto-repair-shop-local-<logical name> (resources.md).
set -eu

PREFIX="auto-repair-shop-local"
REGION="${AWS_REGION:-us-east-1}"
ENDPOINT="${AWS_ENDPOINT_URL:-http://ministack:4566}"
ACCOUNT="000000000000"

aws_local() {
  aws --endpoint-url "$ENDPOINT" --region "$REGION" "$@"
}

# Dead-letter queue for the catalog queue (maxReceiveCount=5, messaging.md).
aws_local sqs create-queue --queue-name "${PREFIX}-catalog-queue-dlq" >/dev/null 2>&1 || true
DLQ_URL="$(aws_local sqs get-queue-url --queue-name "${PREFIX}-catalog-queue-dlq" --query QueueUrl --output text)"
DLQ_ARN="$(aws_local sqs get-queue-attributes --queue-url "$DLQ_URL" --attribute-names QueueArn --query Attributes.QueueArn --output text)"

REDRIVE="{\"deadLetterTargetArn\":\"${DLQ_ARN}\",\"maxReceiveCount\":\"5\"}"
aws_local sqs create-queue --queue-name "${PREFIX}-catalog-queue" \
  --attributes "RedrivePolicy=${REDRIVE}" >/dev/null 2>&1 || true
QUEUE_URL="$(aws_local sqs get-queue-url --queue-name "${PREFIX}-catalog-queue" --query QueueUrl --output text)"
QUEUE_ARN="$(aws_local sqs get-queue-attributes --queue-url "$QUEUE_URL" --attribute-names QueueArn --query Attributes.QueueArn --output text)"

# Topic of the saga commands (shared name, created by whichever service starts first).
TOPIC_ARN="$(aws_local sns create-topic --name "${PREFIX}-saga-commands" --query TopicArn --output text)"

# Let the topic deliver to the queue.
POLICY="{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"sns.amazonaws.com\"},\"Action\":\"sqs:SendMessage\",\"Resource\":\"${QUEUE_ARN}\",\"Condition\":{\"ArnEquals\":{\"aws:SourceArn\":\"${TOPIC_ARN}\"}}}]}"
aws_local sqs set-queue-attributes --queue-url "$QUEUE_URL" --attributes "Policy=${POLICY}"

# Subscription with raw delivery and the target filter for the catalog.
FILTER='{"target":["catalog"]}'
ESCAPED_FILTER="$(printf '%s' "$FILTER" | sed 's/"/\\"/g')"
SUB_ATTRS="{\"RawMessageDelivery\":\"true\",\"FilterPolicy\":\"${ESCAPED_FILTER}\"}"
aws_local sns subscribe --topic-arn "$TOPIC_ARN" --protocol sqs --notification-endpoint "$QUEUE_ARN" \
  --attributes "$SUB_ATTRS" >/dev/null 2>&1 || true

echo "catalog messaging ready: ${QUEUE_URL} (account ${ACCOUNT})"
