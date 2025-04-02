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
      <div class="row user-header__profile-image-container">
        {{template "author-info-no-link" .}}
        <div class="following-info">
          {{template "following-button" .}}
          {{if .IsFollowingYou}}
            <span class="follows-you-label">Follows you</span>
          {{end}}
        </div>
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
        <img class="svg-icon" src="/static/icons/calendar.svg" title="Join date" width="24" height="24" />
        <label>{{.JoinDate.Time.Format "Jan 2, 2006"}}</label>
      </div>

      <div class="followers-followees row">
        <a hx-boost="true" href="/{{.Handle}}/followers" class="followers-followees__followers">
          <span class="followers-followees__count">{{.FollowersCount}}</span>
          <label>followers</label>
        </a>
        <a hx-boost="true" href="/{{.Handle}}/followees" class="followers-followees__followees">
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

      {{if .FollowersYouKnow}}
        <div class="row followers-you-know">
          {{template "N-profile-images" (dict "Users" .FollowersYouKnow "MaxDisplayUsers" 6)}}
          <a hx-boost="true" href="/{{.Handle}}/followers_you_know">
            <span class="followers-you-know__label">...followed by {{(len .FollowersYouKnow)}} you follow</span>
          </a>
        </div>
      {{end}}

      {{if .Lists}}
        <div class="row user-header__lists-container">
          <img class="svg-icon" src="/static/icons/lists.svg" title="Lists this user is on" width="24" height="24" />
          <ul class="user-header__lists">
            {{range .Lists}}
              <li><a href="/lists/{{.ID}}">{{.Name}}</a></li>
            {{end}}
          </ul>
        </div>
      {{end}}
    </div>

    <div class="htmx-spinner">
      <div class="htmx-spinner__fullscreen-forcer">
        <div class="htmx-spinner__background"></div>
        <img class="svg-icon htmx-spinner__icon" src="/static/icons/spinner.svg" />
      </div>
    </div>
  </div>
{{end}}


{{define "N-profile-images"}}
  <div class="N-profile-images" hx-trigger="click consume">
    {{range $i, $user := .Users}}
      {{/* Only render the first N users */}}
      {{if (lt $i $.MaxDisplayUsers)}}
        {{template "circle-profile-img" $user}}
      {{end}}
    {{end}}
  </div>
{{end}}
