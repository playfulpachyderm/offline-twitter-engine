{{define "timeline"}}
  {{range .Items}}
    {{if .NotificationID}}
      {{template "notification" .}}
    {{else}}
      {{template "tweet" .}}
    {{end}}
  {{end}}

  <div class="show-more">
    {{if .CursorBottom.CursorPosition.IsEnd}}
      <label class="show-more__eof-label">End of feed</label>
    {{else}}
      <a class="show-more__button button"
        hx-get="?{{(cursor_to_query_params .CursorBottom)}}"
        hx-target=".show-more"
        hx-swap="outerHTML"
      >Show more</a>
    {{end}}
  </div>
{{end}}
