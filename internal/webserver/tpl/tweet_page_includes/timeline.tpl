{{define "timeline"}}
  {{range .Items}}
    {{template "tweet" .}}
  {{end}}
  {{if .CursorBottom.CursorPosition.IsEnd}}
    <p>End of feed</p>
  {{else}}
    <button class="show-more"
      hx-get="?cursor={{.CursorBottom.CursorValue}}"
      hx-swap="outerHTML"
    >Show more</button>
  {{end}}
{{end}}
