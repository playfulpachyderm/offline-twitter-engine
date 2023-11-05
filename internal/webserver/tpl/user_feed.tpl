{{define "title"}}@{{(user .UserID).Handle}}{{end}}

{{define "main"}}
  {{$user := (user .UserID)}}
  <div class="user-feed-header">
    {{if $user.BannerImageLocalPath}}
      <img class="profile-banner-image" src="/content/profile_images/{{$user.BannerImageLocalPath}}" />
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

        <a class="unstyled-link" target="_blank" href="https://twitter.com/{{$user.Handle}}">
          <li class="quick-link">
            <img class="svg-icon" src="/static/icons/external-link.svg" />
            <span>Open on twitter.com</span>
          </li>
        </a>
        <div class="XXX">
          <a class="unstyled-link" title="Refresh" hx-get="?scrape" hx-target="body">
            <li class="quick-link">
            <img class="svg-icon" src="/static/icons/refresh.svg" />
          </li>
          </a>
        </div>
      </div>
    </div>

    <div class="row user-feed-tabs-container">
      <a class="user-feed-tab unstyled-link {{if (eq .FeedType "")}}active-tab{{end}}" href="/{{$user.Handle}}">
        <span class="user-feed-tab-inner">Tweets and replies</span>
      </a>
      <a class="user-feed-tab unstyled-link {{if (eq .FeedType "without_replies")}}active-tab{{end}}" href="/{{$user.Handle}}/without_replies">
        <span class="user-feed-tab-inner">Tweets</span>
      </a>
      <a class="user-feed-tab unstyled-link {{if (eq .FeedType "media")}}active-tab{{end}}" href="/{{$user.Handle}}/media">
        <span class="user-feed-tab-inner">Media</span>
      </a>
      <a class="user-feed-tab unstyled-link {{if (eq .FeedType "likes")}}active-tab{{end}}" href="/{{$user.Handle}}/likes">
        <span class="user-feed-tab-inner">Likes</span>
      </a>
    </div>
  </div>

  <div class="timeline user-feed-timeline">
    {{template "timeline" .}}
  </div>
{{end}}
