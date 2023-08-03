{{define "author-info"}}
<div class="author-info">
  <a class="unstyled-link" href="/{{.Handle}}">
    <img style="border-radius: 50%; width: 50px; display: inline;" src="{{.ProfileImageUrl}}" />
  </a>
  <span class="name-and-handle">
    <div class="display-name">{{.DisplayName}}</div>
    <div class="handle">@{{.Handle}}</div>
  </span>
</div>
{{end}}
