migrate:
	migrate -path "./migrations" -database "sqlite3://data.db" up
migrate-down:
	migrate -path "./migrations" -database "sqlite3://data.db" down