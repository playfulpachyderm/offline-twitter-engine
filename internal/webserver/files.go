package webserver

import (
	"embed"
)

//go:embed "tpl" "static"
var embedded_files embed.FS

var use_embedded = false
