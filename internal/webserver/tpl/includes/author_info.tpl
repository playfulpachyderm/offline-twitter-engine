{{define "author-info"}}
<div class="author-info" hx-boost="true">
  <a class="unstyled-link" href="/{{.Handle}}">
    <img
      class="profile-image"
      src="/content/{{.GetProfileImageLocalPath}}"
    />
  </a>
  <span class="name-and-handle">
    <div class="display-name row">
      {{.DisplayName}}
      {{if .IsPrivate}}
        <div class="circle-outline">
          <img class="svg-icon" src="/static/icons/lock.svg" />
        </div>
      {{end}}
    </div>
    <div class="handle">@{{.Handle}}</div>
  </span>
</div>
{{end}}
