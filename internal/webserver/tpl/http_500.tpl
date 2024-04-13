{{define "error-toast"}}
  <div class="error-messages__msg" open>
    <span>{{.ErrorMsg}}</span>
    <button class="suicide" onclick="htmx.remove('.error-messages__msg')">X</button>
  </div>
{{end}}
