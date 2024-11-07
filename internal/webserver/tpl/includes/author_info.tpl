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

{{/*
  For User Feed header-- the profile image is no longer a link, but should pop up in the image carousel on click
*/}}
{{define "author-info-no-link"}}
  <div class="author-info" hx-boost="true">
    {{template "circle-profile-img-no-link" .}}
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
