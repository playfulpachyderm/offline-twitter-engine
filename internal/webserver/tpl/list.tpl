{{define "title"}}{{.Title}}{{end}}

{{define "main"}}
  {{template "list" .UserIDs}}
{{end}}
