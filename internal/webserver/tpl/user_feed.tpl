{{define "title"}}@{{(user .UserID).Handle}}{{end}}

{{define "main"}}
  {{$user := (user .UserID)}}
  <div class="user-feed-header">
    {{if $user.BannerImageLocalPath}}
      <img class="profile-banner-image" src="/content/profile_images/{{$user.BannerImageLocalPath}}" />
    {{end}}

    <div class="user-feed-header-info-container">
      {{template "author-info" $user}}
      <button>{{if $user.IsFollowed}}Unfollow{{else}}Follow{{end}}</button>
      <div class="user-bio">
        <span>{{$user.Bio}}</span>
      </div>
      <div class="user-location bio-info-with-icon">
        <img class="svg-icon" src="/static/icons/location.svg" />
        <span>{{$user.Location}}</span>
      </div>
      <div class="user-website bio-info-with-icon">
        <img class="svg-icon" src="/static/icons/website.svg" />
        <span>{{$user.Website}}</span>
      </div>
      <div class="user-join-date bio-info-with-icon">
        <img class="svg-icon" src="/static/icons/calendar.svg" />
        <span>{{$user.JoinDate.Time.Format "Jan 2, 2006"}}</span>
      </div>

      <div class="followers-followees-container">
        <span class="followers-count">{{$user.FollowersCount}}</span>
        <span class="followers-label">followers</span>
        <span class="following-label">is following</span>
        <span class="following-count">{{$user.FollowingCount}}</span>
      </div>
    </div>
  </div>

  <div class="user-feed-tweets">
    {{range .Items}}
      {{template "tweet" .TweetID}}
    {{end}}
  </div>
{{end}}
