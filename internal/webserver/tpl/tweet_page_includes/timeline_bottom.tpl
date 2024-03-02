{{define "timeline-bottom"}}
  <div class="timeline-bottom-container">
    {{if .CursorPosition.IsEnd}}
      <div class="eof-indicator">End of feed</div>
    {{else}}
      <a class="show-more quick-link unstyled-link"
        hx-get="?{{(cursor_to_query_params .)}}"
        hx-target=".timeline-bottom-container"
        hx-swap="outerHTML"
      >Show more</a>
    {{end}}
  </div>
{{end}}
