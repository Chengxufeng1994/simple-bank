# Simple Bank

## How to generate code

* Generate SQL CRUD with sqlc:

```
make sqlc
```

* Generate DB mock with gomock:

```
make mock
```

* Create a new db migration:

```
docker run --rm -v ${PWD}/db/migrations:/migrations --network host migrate/migrate -path=/migrations create -ext sql -dir /migrations -seq <migrations_name>
```