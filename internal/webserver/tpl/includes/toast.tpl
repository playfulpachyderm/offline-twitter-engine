{{define "toast"}}
  <div
    class="toast toast--{{.Type}}"
    {{if .AutoCloseDelay}}
      hx-on::load="setTimeout(() => this.remove(), {{.AutoCloseDelay}} + 2000); setTimeout(() => this.classList.add('disappearing'), {{.AutoCloseDelay}})"
    {{end}}
  >
    <span class="toast__message">{{.Message}}</span>
    {{if not .AutoCloseDelay}}
      <button class="suicide" onclick="this.parentElement.remove()">X</button>
    {{end}}
  </div>
{{end}}
