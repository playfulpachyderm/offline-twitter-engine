{{define "author-info"}}
<div class="author-info">
  <a class="unstyled-link" href="/{{.Handle}}">
    <img class="profile-image" src="{{if .IsContentDownloaded}}/content/profile_images/{{.ProfileImageLocalPath}}{{else}}{{.ProfileImageUrl}}{{end}}" />
  </a>
  <span class="name-and-handle">
    <div class="display-name">{{.DisplayName}}</div>
    <div class="handle">@{{.Handle}}</div>
  </span>
</div>
{{end}}
