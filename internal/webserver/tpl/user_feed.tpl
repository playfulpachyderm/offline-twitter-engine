{{define "title"}}@{{(user .UserID).Handle}}{{end}}

{{define "main"}}
  {{$user := (user .UserID)}}
  <div class="user-feed-header">
    {{if $user.BannerImageLocalPath}}
      {{if $user.IsContentDownloaded}}
        <img class="profile-banner-image" src="/content/profile_images/{{$user.BannerImageLocalPath}}" />
      {{else}}
        <img class="profile-banner-image" src="{{$user.BannerImageUrl}}" />
      {{end}}
    {{end}}

    <div class="user-feed-header-info-container">
      <div class="row">
        {{template "author-info" $user}}
        {{template "following-button" $user}}
      </div>
      <div class="user-bio">
        {{template "text-with-entities" $user.Bio}}
      </div>
      {{if $user.Location}}
        <div class="user-location bio-info-with-icon">
          <img class="svg-icon" src="/static/icons/location.svg" />
          <span>{{$user.Location}}</span>
        </div>
      {{end}}
      {{if $user.Website}}
        <div class="user-website bio-info-with-icon">
          <img class="svg-icon" src="/static/icons/website.svg" />
          <a class="unstyled-link" target="_blank" href="{{$user.Website}}">{{$user.Website}}</a>
        </div>
      {{end}}
      <div class="user-join-date bio-info-with-icon">
        <img class="svg-icon" src="/static/icons/calendar.svg" />
        <span>{{$user.JoinDate.Time.Format "Jan 2, 2006"}}</span>
      </div>

      <div class="followers-followees-container row">
        <div class="followers-container">
          <span class="followers-count">{{$user.FollowersCount}}</span>
          <span class="followers-label">followers</span>
        </div>
        <div class="followees-container">
          <span class="following-label">is following</span>
          <span class="following-count">{{$user.FollowingCount}}</span>
        </div>

        <div class="spacer"></div>

        <div class="user-feed-buttons-container">
          <a class="unstyled-link quick-link" target="_blank" href="https://twitter.com/{{$user.Handle}}" title="Open on twitter.com">
            <img class="svg-icon" src="/static/icons/external-link.svg" />
          </a>
          <a class="unstyled-link quick-link" hx-get="?scrape" hx-target="body" title="Refresh">
            <img class="svg-icon" src="/static/icons/refresh.svg" />
          </a>
        </div>
      </div>
    </div>

    <div class="row tabs-container">
      <a class="tab unstyled-link {{if (eq .FeedType "")}}active-tab{{end}}" href="/{{$user.Handle}}">
        <span class="tab-inner">Tweets and replies</span>
      </a>
      <a class="tab unstyled-link {{if (eq .FeedType "without_replies")}}active-tab{{end}}" href="/{{$user.Handle}}/without_replies">
        <span class="tab-inner">Tweets</span>
      </a>
      <a class="tab unstyled-link {{if (eq .FeedType "media")}}active-tab{{end}}" href="/{{$user.Handle}}/media">
        <span class="tab-inner">Media</span>
      </a>
      <a class="tab unstyled-link {{if (eq .FeedType "likes")}}active-tab{{end}}" href="/{{$user.Handle}}/likes">
        <span class="tab-inner">Likes</span>
      </a>
    </div>
  </div>

  <div class="timeline user-feed-timeline">
    {{template "timeline" .}}
  </div>
{{end}}
