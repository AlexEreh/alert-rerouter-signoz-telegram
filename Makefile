.PHONY: env-up env-down

env-up:
	docker compose up -d --build

env-down:
	docker compose down