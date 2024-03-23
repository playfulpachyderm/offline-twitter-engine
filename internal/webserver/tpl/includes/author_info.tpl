{{define "author-info"}}
<div class="author-info" hx-boost="true">
  <a class="unstyled-link" href="/{{.Handle}}">
    <img
      class="profile-image"
      {{if .IsContentDownloaded}}
        src="/content/{{.GetProfileImageLocalPath}}"
      {{else}}
        src="{{.ProfileImageUrl}}"
      {{end}}
    />
  </a>
  <span class="name-and-handle">
    <div class="display-name row">
      {{.DisplayName}}
      {{if .IsPrivate}}
        <div class="circle-outline">
          <img class="svg-icon" src="/static/icons/lock.svg" width="24" height="24" />
        </div>
      {{end}}
    </div>
    <div class="handle">@{{.Handle}}</div>
  </span>
</div>
{{end}}
