{{define "title"}}Messages{{end}}

{{define "main"}}
  <div class="chats-container">
    {{template "chat-list" .}}
    {{template "chat-view" .}}
  </div>
  <div class="spacer"></div>
{{end}}
