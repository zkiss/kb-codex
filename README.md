# kb-codex

## Usage

```sh
go run ./cmd/server --addr ":8080"
```

The server will start at the specified address (default is `:8080`).

## Knowledge Bases

- Create and list knowledge bases, upload `.txt`/`.md` files, and store chunk embeddings in the vector DB.
- A single-page React UI is available at the root path. It uses semantic routes
  like `/login`, `/register`, `/kbs`, and `/kbs/<kb-id>` when viewing a specific
  knowledge base.

### API Endpoints

| Method | Path                         | Description                               |
|--------|------------------------------|-------------------------------------------|
| GET    | `/api/kbs`                   | List all knowledge bases                  |
| POST   | `/api/kbs`                   | Create a new knowledge base (`{name}`)    |
| GET    | `/api/kbs/{kbID}/files`      | List uploaded files in a KB               |
| POST   | `/api/kbs/{kbID}/files`      | Upload `.txt`/`.md` file and index chunks |
| POST   | `/api/kbs/{kbID}/ask`        | Ask a question about a KB (`{question}`) |

Set the `OPENAI_API_KEY` environment variable to enable embeddings.

Migrations are applied automatically on startup (using `./migrations`).