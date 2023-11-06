{{define "title"}}Followed Users{{end}}

{{define "main"}}
<div class="users-list-container">
  {{range .}}
    {{template "author-info" .}}
  {{end}}
</div>
{{end}}
