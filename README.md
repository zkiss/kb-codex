# kb-codex

## Usage

```sh
go run ./cmd/server --addr ":8080"
```

The server will start at the specified address (default is `:8080`).

### Frontend development

The React UI lives under `ui/` and is built with [Vite](https://vitejs.dev/).
During development the Go server should run on `:8080` so the Vite dev server
can proxy API requests.

```sh
cd ui
npm install
npm run dev   # start Vite dev server (proxies /api to :8080)
npm run build # output production assets into ../static
```

Run `npm run build` before starting the Go server to populate the `static/`
directory with the production assets. The production build expects the
assets to be served from the `/static` path, which the Go server is
configured to provide.

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