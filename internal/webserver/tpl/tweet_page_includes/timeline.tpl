{{define "timeline"}}
  {{range .Items}}
    {{template "tweet" .}}
  {{end}}

  <div class="timeline__bottom">
    {{if .CursorBottom.CursorPosition.IsEnd}}
      <label class="timeline__eof-label">End of feed</label>
    {{else}}
      <a class="timeline__show-more-button button"
        hx-get="?{{(cursor_to_query_params .CursorBottom)}}"
        hx-target=".timeline__bottom"
        hx-swap="outerHTML"
      >Show more</a>
    {{end}}
  </div>
{{end}}
