POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= postgres
POSTGRES_URL ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/simple_bank?sslmode=disable

postgres:
	docker run --name postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -d -p 5432:5432 --restart=always postgres:12.11

createdb:
	docker exec -it postgres createdb --username=postgres --owner=postgres simple_bank

dropdb:
	docker exec -it postgres dropdb -U postgres simple_bank

migrate_up:
	docker run --rm -v ${PWD}/db/migrations:/migrations --network host migrate/migrate -path=/migrations -database $(POSTGRES_URL) -verbose up

migrate_down:
	docker run --rm -v ${PWD}/db/migrations:/migrations --network host migrate/migrate -path=/migrations -database $(POSTGRES_URL) -verbose down -all

migrate_version:
	docker run --rm -v ${PWD}/db/migrations:/migrations --network host migrate/migrate -path=/migrations -database $(POSTGRES_URL) -verbose version

sqlc:
	docker run --rm -v ${PWD}:/src -w /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/Chengxufeng1994/simple-bank/db/sqlc Store

db_docs:
	npm run db_docs

db_schema:
	npm run db_schema

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
		proto/*.proto

evans:
	evans --host localhost --port 9090 -r repl

.PHONY: postgres createdb dropdb migrate_up migrate_down migrate_version sqlc mock test server db_docs db_schema proto evans

SHELL = /bin/sh