{{define "title"}}{{.Title}}{{end}}

{{define "main"}}
  {{template "list" .Users}}
{{end}}
