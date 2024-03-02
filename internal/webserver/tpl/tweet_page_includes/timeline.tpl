{{define "timeline"}}
  {{range .Items}}
    {{template "tweet" .}}
  {{end}}

  {{template "timeline-bottom" .CursorBottom}}
{{end}}
