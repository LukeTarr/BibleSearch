package docs

import "embed"

// FS holds the hand-written OpenAPI spec and the Swagger UI page, served at /swagger/
//
//go:embed index.html openapi.yaml
var FS embed.FS
