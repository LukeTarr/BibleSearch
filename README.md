# BibleSearch

**[biblesearch.ing](https://biblesearch.ing)**: search the Bible by meaning, not keywords.

Normal Bible search needs you to remember the exact wording. BibleSearch lets you describe what you're
looking for in plain language ("Eve eats the apple", "don't worry about tomorrow") and returns the verses
closest in *meaning*, even when they share no words with your query. It uses the King James Version (all
66 books, 31,100 verses), and every result links to the chapter on bible.com.

## How it works

It's a semantic search built on vector embeddings:

1. **Vectorize (one time):** every verse in `data/en_kjv.json` is sent to OpenAI's `text-embedding-3-small`
   model in batches of 1,000, which turns each one into a 1536-number vector that captures its meaning.
   Chroma, a vector database, stores those vectors along with the book, chapter and verse number.
2. **Search:** your query is embedded with the same model, and Chroma returns the 10 verses whose vectors
   are nearest to it (L2 distance). Near vectors mean similar meaning.

The query and the stored verses must be embedded with the same model. If you change `EmbeddingModel` in
`services/chroma.go`, the collection name changes with it, and you have to vectorize again.

## Tech stack

- **Go + [Gin](https://gin-gonic.com/)**: HTTP server
- **[templ](https://templ.guide/)**: type-safe HTML templates (`templates/*.templ` → generated `*_templ.go`)
- **[htmx](https://htmx.org/)**: the search box posts to `/search` and swaps in the results HTML (no JS framework)
- **Tailwind + Flowbite**: styling, loaded from CDNs in `templates/header.templ`
- **[Chroma](https://www.trychroma.com/)** 1.x: vector database, called through its v2 REST API with plain
  `net/http`. OpenAI embeddings are called the same way. No client libraries.
- **Swagger** (`swaggo`): API docs at `/swagger/index.html`

## Project layout

```
main.go              startup: config, Chroma collection, routes
controllers/         route registration (pages + /api/v1)
services/
  chroma.go          collection setup, embedding model, query + HTMX handlers
  chromaclient.go    minimal Chroma v2 REST client
  openai.go          OpenAI embeddings client
  vectorization.go   loads the Bible and embeds every verse into Chroma
  config.go          env vars / .env loading
  data.go            parses data/en_kjv.json
templates/           templ components (home, about, header)
data/                KJV text + book → bible.com abbreviation map
docs/                generated Swagger spec (don't edit by hand)
```

## Getting started

**Prerequisites:** Go 1.21+, Docker, an OpenAI API key, and these tools:

```sh
go install github.com/a-h/templ/cmd/templ@v0.2.778   # must match the templ version in go.mod
go install github.com/air-verse/air@latest           # live reload
go install github.com/swaggo/swag/cmd/swag@latest    # swagger generation
```

**1. Start Chroma** (port 8000, data kept in a Docker volume):

```sh
docker compose up -d
```

**2. Create `.env`** in the repo root (it's gitignored):

```sh
CHROMA_URL=http://localhost:8000
OPENAI_API_KEY=sk-...
VECTORIZATION_PASSWORD=pick-something
```

**3. Run the server** with live reload at http://localhost:8080:

```sh
./dev.sh
```

After editing a `.templ` file, run `templ generate`. The generated `*_templ.go` files are committed.

**4. Load the verses** (one time per Chroma volume). Open http://localhost:8080/swagger/index.html and call
`POST /api/v1/vectorize` with `{"password": "<VECTORIZATION_PASSWORD>"}`. It runs in the background: about
30 batched embedding calls, which take a minute or two and cost a couple of cents. The server logs
`Counted documents` when it's done. Don't start it twice at the same time.

## API

| Route | Purpose |
|---|---|
| `GET /`, `GET /about` | Pages |
| `POST /search` | htmx search (form field `query`), returns HTML |
| `POST /api/v1/query` | JSON search: `{"query": "..."}` → top 10 verses with distances |
| `POST /api/v1/vectorize` | Embed every verse into Chroma (password protected) |

## Deployment

Production runs on a VPS with [Coolify](https://coolify.io/), which builds the `Dockerfile` on every push to
`main`. Chroma runs as a separate container, pinned to the same image tag as `docker-compose.yml` (data at
`/data`, `CHROMA_ALLOW_RESET=true`). Secrets come from Coolify's environment variables at runtime.
No `.env` file ends up in the image, and the app falls back to the environment when the file isn't there.
The templ generator image in the `Dockerfile` is pinned to match `go.mod`. Bump both together.
