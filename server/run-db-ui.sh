podman run --rm --name pgadmin \
  -p 8081:8080 \
  -e PGADMIN_DEFAULT_EMAIL=admin@admin.com \
  -e PGADMIN_DEFAULT_PASSWORD=admin \
  -e PGADMIN_LISTEN_PORT=8080 \
  docker.io/dpage/pgadmin4