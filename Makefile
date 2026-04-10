migrate:
	migrate -path "./migrations" -database "sqlite3://data.db" up