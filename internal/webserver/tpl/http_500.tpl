{{define "error-toast"}}
  <div class="server-error-msg">
    <div class="error-msg-container">
      <span>{{.ErrorMsg}}</span>
      <button class="suicide" onclick="htmx.remove('.server-error-msg')">X</button>
    </div>
  </div>
{{end}}
