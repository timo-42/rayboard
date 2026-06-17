package docs

import "embed"

// Files contains the project documentation markdown served by the frontend.
//
//go:embed *.md
var Files embed.FS
