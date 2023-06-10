DB_URL=postgresql://mars:mars@localhost:5555/storage-king?sslmode=disable
postgres:
	docker run --name postgres14 --network bank-network -p 5555:5432 -e POSTGRES_USER=mars -e POSTGRES_PASSWORD=mars -d postgres:14

start_postgres:
	docker start postgres14

stop_postgres:
	docker stop postgres14

createdb:
	docker exec -it postgres14 createdb --username=mars --owner=mars storage-king

dropdb:
	docker exec -it postgres14 dropdb --username=mars storage-king

migrateup:
	migrate -path db/migrations -database "${DB_URL}" -verbose up

migratedown:
	migrate -path db/migrations -database "${DB_URL}" -verbose down

forcing:
	migrate -path db/migrations -database "${DB_URL}" force $(version)

new_migration:
	migrate create -ext sql -dir db/migrations -seq $(name)

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate

server:
	go run main.go

.PHONY: postgres start_postgres stop_postgres createdb dropdb migrateup migratedown new_migration sqlc server db_docs db_schema