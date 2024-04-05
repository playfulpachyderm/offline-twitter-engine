{{define "title"}}Messages{{end}}

{{define "main"}}
  <div class="messages-page">
    {{template "chat-list" .}}
    {{template "chat-view" .}}
  </div>
  <div class="spacer"></div>
{{end}}
