// Package docs menyematkan spesifikasi OpenAPI agar ikut ter-build ke binary.
package docs

import _ "embed"

//go:embed openapi.yaml
var OpenAPI []byte
