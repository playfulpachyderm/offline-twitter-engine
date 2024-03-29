{{define "user-header"}}
  <div class="user-header">
    {{if .BannerImageLocalPath}}
      {{if .IsContentDownloaded}}
        <img class="profile-banner-image" src="/content/profile_images/{{.BannerImageLocalPath}}" />
      {{else}}
        <img class="profile-banner-image" src="{{.BannerImageUrl}}" />
      {{end}}
    {{end}}

    <div class="user-header-info-container">
      <div class="row">
        {{template "author-info" .}}
        {{template "following-button" .}}
      </div>
      <div class="user-bio">
        {{template "text-with-entities" .Bio}}
      </div>
      {{if .Location}}
        <div class="user-location bio-info-with-icon">
          <img class="svg-icon" src="/static/icons/location.svg" width="24" height="24" />
          <span>{{.Location}}</span>
        </div>
      {{end}}
      {{if .Website}}
        <div class="user-website bio-info-with-icon">
          <img class="svg-icon" src="/static/icons/website.svg" width="24" height="24" />
          <a class="unstyled-link" target="_blank" href="{{.Website}}">{{.Website}}</a>
        </div>
      {{end}}
      <div class="user-join-date bio-info-with-icon">
        <img class="svg-icon" src="/static/icons/calendar.svg" width="24" height="24" />
        <span>{{.JoinDate.Time.Format "Jan 2, 2006"}}</span>
      </div>

      <div class="followers-followees-container row">
        <a href="/{{.Handle}}/followers" class="followers-container unstyled-link">
          <span class="followers-count">{{.FollowersCount}}</span>
          <span class="followers-label">followers</span>
        </a>
        <a href="/{{.Handle}}/followees" class="followers-container unstyled-link">
          <span class="following-label">is following</span>
          <span class="following-count">{{.FollowingCount}}</span>
        </a>

        <div class="spacer"></div>

        <div class="user-feed-buttons-container">
          <a class="unstyled-link quick-link" target="_blank" href="https://twitter.com/{{.Handle}}" title="Open on twitter.com">
            <img class="svg-icon" src="/static/icons/external-link.svg" width="24" height="24" />
          </a>
          <a class="unstyled-link quick-link" hx-get="?scrape" hx-target="body" hx-indicator=".user-header" title="Refresh">
            <img class="svg-icon" src="/static/icons/refresh.svg" width="24" height="24" />
          </a>
        </div>
      </div>
    </div>

    <div class="htmx-spinner-container">
      <div class="htmx-spinner-background"></div>
      <img class="svg-icon htmx-spinner" src="/static/icons/spinner.svg" />
    </div>
  </div>
{{end}}
