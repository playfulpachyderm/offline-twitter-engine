{{define "following-button"}}
  {{ $action := "follow" }}
  {{if .IsFollowed}}
    {{ $action = "unfollow" }}
  {{end}}

  <div class="button following-button"
    hx-post="/{{$action}}/{{.Handle}}"
    hx-swap="outerHTML"
    style="text-transform: capitalize; position: relative"
  >
    <div class="htmx-spinner">
      <div class="htmx-spinner__background"></div>
      <img class="svg-icon htmx-spinner__icon" src="/static/icons/spinner.svg" />
    </div>

    {{$action}}
  </div>
{{end}}
