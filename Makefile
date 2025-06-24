postgres:
	docker run --name postgres17.5 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:17.5-alpine

createdb:
	 docker exec -it postgres17.5 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres17.5 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mockg:
	mockgen -package mockdb -destination db/mock/store.go github.com/avfirsov/golang-backend-masterclass/db/sqlc Store

.PHONY: createdb dropdb postgres migrateup migratedown migratedown1 migrateup1 sqlc test server mock