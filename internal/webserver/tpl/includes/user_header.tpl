{{define "user-header"}}
  <div class="user-header">
    {{if .BannerImageLocalPath}}
      <img class="user-header__profile-banner-image"
        onclick="image_carousel.querySelector('img').src = this.src; image_carousel.showModal();"
        {{if .IsContentDownloaded}}
          src="/content/profile_images/{{.BannerImageLocalPath}}"
        {{else}}
          src="{{.BannerImageUrl}}"
        {{end}}
      >
    {{end}}

    <div class="user-header__info-container">
      <div class="row">
        {{template "author-info-no-link" .}}
        {{template "following-button" .}}
      </div>
      <div class="user-header__bio">
        {{template "text-with-entities" .Bio}}
      </div>
      {{if .Location}}
        <div class="user-header__location labelled-icon">
          <img class="svg-icon" src="/static/icons/location.svg" width="24" height="24" />
          <label>{{.Location}}</label>
        </div>
      {{end}}
      {{if .Website}}
        <div class="user-header__website labelled-icon">
          <img class="svg-icon" src="/static/icons/website.svg" width="24" height="24" />
          <label><a target="_blank" href="{{.Website}}">{{.Website}}</a></label>
        </div>
      {{end}}
      <div class="user-header__join-date labelled-icon">
        <img class="svg-icon" src="/static/icons/calendar.svg" width="24" height="24" />
        <label>{{.JoinDate.Time.Format "Jan 2, 2006"}}</label>
      </div>

      <div class="followers-followees row">
        <a href="/{{.Handle}}/followers" class="followers-followees__followers">
          <span class="followers-followees__count">{{.FollowersCount}}</span>
          <label>followers</label>
        </a>
        <a href="/{{.Handle}}/followees" class="followers-followees__followees">
          <label>is following</label>
          <span class="followers-followees__count">{{.FollowingCount}}</span>
        </a>

        <div class="spacer"></div>

        <div class="row">
          <a class="button" target="_blank" href="https://twitter.com/{{.Handle}}" title="Open on twitter.com">
            <img class="svg-icon" src="/static/icons/external-link.svg" width="24" height="24" />
          </a>
          <a class="button" hx-get="?scrape" hx-target="body" hx-indicator=".user-header" title="Refresh">
            <img class="svg-icon" src="/static/icons/refresh.svg" width="24" height="24" />
          </a>
        </div>
      </div>
    </div>

    <div class="htmx-spinner">
      <div class="htmx-spinner__fullscreen-forcer">
        <div class="htmx-spinner__background"></div>
        <img class="svg-icon htmx-spinner__icon" src="/static/icons/spinner.svg" />
      </div>
    </div>
  </div>
{{end}}
