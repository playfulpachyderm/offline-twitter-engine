{{define "following-button"}}
  {{if .IsFollowed}}
    <button class="following-button"
      hx-post="/unfollow/{{.Handle}}"
      hx-swap="outerHTML"
    >
      Unfollow
    </button>
  {{else}}
    <button class="following-button"
      hx-post="/follow/{{.Handle}}"
      hx-swap="outerHTML"
    >
      Follow
    </button>
  {{end}}
{{end}}
