#!/bin/bash

set -e

BOT_TOKEN=${BOT_TOKEN:?"BOT_TOKEN must be specified"}
WEBHOOK_FUNC_NAME=${WEBHOOK_FUNC_NAME:-"tg-webhook-updates"}
WEBHOOK_FUNC_REGION=${WEBHOOK_FUNC_REGION:-"europe-west2"}
WEBHOOK_TOKEN=${WEBHOOK_TOKEN:-$(openssl rand -hex 64)}

curl -X POST -F "drop_pending_updates=True" https://api.telegram.org/bot"${BOT_TOKEN}"/deleteWebhook \
; echo

gcloud functions deploy "${WEBHOOK_FUNC_NAME}" \
  --runtime go116 \
  --trigger-http \
  --allow-unauthenticated \
  --entry-point=WebhookHandler \
  --set-env-vars BOT_TOKEN="${BOT_TOKEN}",WEBHOOK_TOKEN="${WEBHOOK_TOKEN}" \
  --memory=128MB \
  --region="${WEBHOOK_FUNC_REGION}" \
; echo

URL=${URL:-$(gcloud functions describe --region="${WEBHOOK_FUNC_REGION}" "${WEBHOOK_FUNC_NAME}" --format="value(httpsTrigger.url)")}

curl \
  -X POST \
  -F "url=${URL}/${WEBHOOK_TOKEN}" \
  -F 'allowed_updates=["message", "edited_message", "inline_query"]' https://api.telegram.org/bot"${BOT_TOKEN}"/setWebhook \
; echo