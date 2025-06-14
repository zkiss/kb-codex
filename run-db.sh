podman run --rm --name local-postgres \
  -e POSTGRES_USER=demo               \
  -e POSTGRES_PASSWORD=demo_pw        \
  -e PGDATA=/var/lib/postgresql/data  \
  -p 5432:5432                        \
  -v "$(pwd)/db.local:/var/lib/postgresql/data:Z" \
  docker.io/pgvector/pgvector:pg17