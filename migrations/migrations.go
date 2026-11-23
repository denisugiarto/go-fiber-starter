// Package migrations menyematkan berkas SQL migrasi (urut menurut nama berkas).
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
