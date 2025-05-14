**To start tests:**

cd to either forum or auth service
go test -v ./...

**To start the services**

cd to either
go run cmd/main.go

**Migrations:**

**template**

migrate -path auth-service/migrations -database "postgres://user:password@localhost:5432/database?sslmode=disable" up
**actually works**

migrate -path auth-service/migrations -database "postgres://postgres:1437@localhost:5432/auth_db?sslmode=disable" up