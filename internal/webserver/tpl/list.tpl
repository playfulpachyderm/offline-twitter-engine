{{define "title"}}{{.Title}}{{end}}

{{define "main"}}
  {{if .HeaderUserID}}
    {{template "user-header" (user .HeaderUserID)}}
  {{else if .HeaderTweetID}}
    {{template "tweet" (tweet .HeaderTweetID)}}
  {{end}}

  <h3>
    {{.Title}}
  </h3>

  {{template "list" .UserIDs}}
{{end}}
