{{define "title"}}Bookmarks{{end}}

{{define "main"}}
  <div class="bookmarks-feed-header">
    <h1>Bookmarks</h1>
  </div>
  <div class="timeline">
    {{template "timeline" .Feed}}
  </div>
{{end}}
