{{define "likes-count"}}
  <div class="interactions__stat" hx-trigger="click consume">
    {{if .IsLikedByCurrentUser}}
      <img class="svg-icon interactions__like-icon interactions__like-icon--liked" src="/static/icons/like_filled.svg" width="24" height="24"
        hx-get="/tweet/{{.ID}}/unlike"
        hx-target="closest .interactions__stat"
        hx-push-url="false"
        hx-swap="outerHTML focus-scroll:false"
      />
    {{else}}
      <img class="svg-icon interactions__like-icon" src="/static/icons/like.svg" width="24" height="24"
        hx-get="/tweet/{{.ID}}/like"
        hx-target="closest .interactions__stat"
        hx-push-url="false"
        hx-swap="outerHTML focus-scroll:false"
      />
    {{end}}
    <span>{{.NumLikes}}</span>
  </div>
{{end}}

{{define "retweets-count"}}
  <div class="interactions__stat" hx-trigger="click consume">
    {{if .IsRetweetedByCurrentUser}}
      <img class="svg-icon interactions__retweet-icon interactions__retweet-icon--retweeted"
        src="/static/icons/retweet.svg" width="24" height="24"
      />
    {{else}}
      <img class="svg-icon interactions__retweet-icon"
        src="/static/icons/retweet.svg" width="24" height="24"
      />
    {{end}}
    <span>{{.NumRetweets}}</span>
  </div>
{{end}}
