BOT_TOKEN?=<your_bot_token>

WEBHOOK_FUNC_NAME?=tg-webhook-updates
WEBHOOK_FUNC_REGION?=europe-west2

.PHONY: deploy wh_func wh_set wh_del

deploy: export WEBHOOK_TOKEN?=$(shell openssl rand -hex 64)
deploy:
	$(MAKE) wh_del wh_func wh_set

wh_func:
	gcloud functions deploy ${WEBHOOK_FUNC_NAME} \
		--runtime go116 \
		--trigger-http \
		--allow-unauthenticated \
		--entry-point=WebhookUpdatesHandler \
		--set-env-vars BOT_TOKEN=${BOT_TOKEN},WEBHOOK_TOKEN=${WEBHOOK_TOKEN} \
		--memory=128MB \
		--region=${WEBHOOK_FUNC_REGION} \
	; echo

wh_set: export URL?=$(shell gcloud functions describe --region=${WEBHOOK_FUNC_REGION} ${WEBHOOK_FUNC_NAME} --format="value(httpsTrigger.url)")
wh_set:
	curl \
		-X POST \
		-F "url=${URL}/${WEBHOOK_TOKEN}" \
		-F 'allowed_updates=["message", "edited_message", "inline_query"]' https://api.telegram.org/bot${BOT_TOKEN}/setWebhook \
	; echo

wh_del:
	curl -X POST -F "drop_pending_updates=True" https://api.telegram.org/bot${BOT_TOKEN}/deleteWebhook \
	; echo