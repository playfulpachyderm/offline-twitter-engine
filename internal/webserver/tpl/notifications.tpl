{{define "title"}}Notifications{{end}}

{{define "main"}}
  <div class="notifications-header">
    <h2>Notifications</h2>
  </div>

  <div class="timeline">
    {{template "timeline" .}}
  </div>
{{end}}
