{{define "author-info"}}
<div class="author-info" hx-boost="true">
  {{template "circle-profile-img" .}}
  <span class="author-info__name-and-handle">
    <div class="author-info__display-name row">
      {{.DisplayName}}
      {{if .IsPrivate}}
        <div class="circle-outline">
          <img class="svg-icon" src="/static/icons/lock.svg" width="24" height="24">
        </div>
      {{end}}
    </div>
    <div class="author-info__handle">@{{.Handle}}</div>
  </span>
</div>
{{end}}
