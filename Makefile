include .env

db-up:
	@migrate -path=./migrations -database=$(DB_DSN) up

db-down:
	@migrate -path=./migrations -database=$(DB_DSN) down

db-migration:
	@migrate create -ext=sql -dir=./migrations $(name)

app-test:
	@migrate -path=./migrations -database=$(DB_DSN_TEST) up
	@go test ./... -v
	@yes | migrate -path=./migrations -database=$(DB_DSN_TEST) down

app-build:
	@go build -o server ./cmd/api

app-run: app-build
	./server

docker-build:
	@docker build -t showtime-api .

compose-up: docker-build
	@docker compose up


psql:
	@psql -U postgres --dbname=$(DB_DATABASE)

info:
	@echo dsn=$(DB_DSN)

swagger:
	swag init --generalInfo cmd/api/main.go --output cmd/api/docs --dir .
