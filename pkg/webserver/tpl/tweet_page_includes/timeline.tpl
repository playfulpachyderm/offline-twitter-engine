{{define "timeline"}}
  {{range .Items}}
    {{if .NotificationID}}
      {{template "notification" .}}
    {{else}}
      {{template "tweet" .}}
    {{end}}
  {{end}}

  <div class="show-more" style="position: relative">
    {{if .CursorBottom.CursorPosition.IsEnd}}
      <label class="show-more__eof-label">End of feed</label>
    {{else}}
      <a class="show-more__button button"
        hx-get="?{{(cursor_to_query_params .CursorBottom)}}"
        hx-target=".show-more"
        hx-swap="outerHTML"
        hx-indicator="closest .show-more"
      >Show more</a>
    {{end}}

    {{/* Set the font-size to override the default style (3em) to prevent it overflowing the box */}}
    <div class="htmx-spinner" style="font-size: 2em">
      <div class="htmx-spinner__background"></div>
      <img class="svg-icon htmx-spinner__icon" src="/static/icons/spinner.svg" />
    </div>
  </div>
{{end}}
