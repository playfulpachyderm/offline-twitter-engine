{{define "likes-count"}}
  <div class="interaction-stat" hx-trigger="click consume">
    {{if .IsLikedByCurrentUser}}
      <img class="svg-icon like-icon liked" src="/static/icons/like_filled.svg" width="24" height="24"
        hx-get="/tweet/{{.ID}}/unlike"
        hx-target="closest .interaction-stat"
        hx-push-url="false"
        hx-swap="outerHTML focus-scroll:false"
      />
    {{else}}
      <img class="svg-icon like-icon" src="/static/icons/like.svg" width="24" height="24"
        hx-get="/tweet/{{.ID}}/like"
        hx-target="closest .interaction-stat"
        hx-push-url="false"
        hx-swap="outerHTML focus-scroll:false"
      />
    {{end}}
    <span>{{.NumLikes}}</span>
  </div>
{{end}}
