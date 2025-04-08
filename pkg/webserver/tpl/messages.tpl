{{define "main"}}
  <div class="messages-page">
    <script type="module" src="/static/vendor/emoji-picker/picker.js"></script>
    {{template "chat-list" .}}
    {{template "chat-view" .}}
  </div>
  <div class="spacer"></div>
{{end}}
