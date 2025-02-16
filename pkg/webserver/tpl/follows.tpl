{{define "title"}}{{.Title}}{{end}}

{{define "main"}}
  {{template "user-header" (user .HeaderUserID)}}

  <h1>
    {{.Title}}
  </h1>

  {{template "list" (dict "UserIDs" .UserIDs)}}
{{end}}
