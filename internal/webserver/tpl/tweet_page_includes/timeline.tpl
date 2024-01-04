{{define "timeline"}}
  {{range .Items}}
    {{template "tweet" .}}
  {{end}}
  {{if .CursorBottom.CursorPosition.IsEnd}}
    <div class="eof-indicator">End of feed</div>
  {{else}}
    <button class="show-more"
      hx-get="?{{(cursor_to_query_params .CursorBottom)}}"
      hx-swap="outerHTML"
    >Show more</button>
  {{end}}
{{end}}
