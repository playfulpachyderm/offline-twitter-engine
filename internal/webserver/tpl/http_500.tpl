{{define "error-toast"}}
  <dialog class="server-error-msg" open>
    <span>{{.ErrorMsg}}</span>
    <button class="suicide" onclick="htmx.remove('.server-error-msg')">X</button>
  </dialog>
{{end}}
