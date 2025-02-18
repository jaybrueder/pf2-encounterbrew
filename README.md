# PF2 Encounterbrew

## Development Dependencies

- pre-commit
- golangci-lint
- docker
- docker-compose
- npm install -g tailwindcss@3.4.17

```
$ docker compose --profile dev up -d

$ export PF2ENCOUNTERBREW_DB_DSN=postgres://admin:admin@localhost/encounterbrew?sslmode=disable

$ migrate -path=./migrations -database=$PF2ENCOUNTERBREW_DB_DSN up

$ go run cmd/seed/seeder.go
```

## Production

```
docker compose --profile prod up -d
```
