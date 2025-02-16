package webserver

import (
	"embed"
	"io/fs"
	"path"
	"path/filepath"
	"runtime"
)

//go:embed "tpl" "static"
var embedded_files embed.FS

var use_embedded = ""

var this_dir string

func init() {
	_, this_file, _, _ := runtime.Caller(0) // `this_file` is absolute path to this source file
	this_dir = path.Dir(this_file)
}

func get_filepath(s string) string {
	if use_embedded == "true" {
		return s
	}
	return path.Join(this_dir, s)
}

func glob(path string) []string {
	var ret []string
	var err error
	if use_embedded == "true" {
		ret, err = fs.Glob(embedded_files, get_filepath(path))
	} else {
		ret, err = filepath.Glob(get_filepath(path))
	}
	panic_if(err)
	return ret
}
