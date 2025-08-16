. ./env.local.sh
export DATABASE_URL=postgres://demo:demo_pw@localhost:5432/postgres?sslmode=disable
export JWT_SECRET=secret
go run ./cmd/server