{{define "title"}}@{{(user .UserID).Handle}}{{end}}

{{define "main"}}
  {{$user := (user .UserID)}}
  <img class="profile-banner-image" src="{{$user.BannerImageUrl}}" />

  {{template "author-info" $user}}
  <button>{{if $user.IsFollowed}}Unfollow{{else}}Follow{{end}}</button>
  <p class="user-bio">{{$user.Bio}}</p>
  <p class="user-location">{{$user.Location}}</p>
  <p class="user-website">{{$user.Website}}</p>
  <p class="user-join-date">{{$user.JoinDate}}</p>

  <div class="followers-followees-container">
    <span class="followers-count">{{$user.FollowersCount}}</span>
    <span class="followers-label">followers</span>
    <span class="following-label">is following</span>
    <span class="following-count">{{$user.FollowingCount}}</span>
  </div>

  <hr/>

  {{range .Items}}
    {{template "tweet" .TweetID}}
  {{end}}
{{end}}
