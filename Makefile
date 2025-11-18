include .env

db-up:
	@migrate -path=./migrations -database=$(DB_DSN) up

db-down:
	@migrate -path=./migrations -database=$(DB_DSN) down

db-migration:
	@migrate create -ext=sql -dir=./migrations $(name)

psql:
	@psql -U postgres --dbname=$(DB_DATABASE)

info:
	@echo dsn=$(DB_DSN)

swagger:
	swag init --generalInfo cmd/api/main.go --output cmd/api/docs --dir .
