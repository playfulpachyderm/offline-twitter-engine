{{define "title"}}@{{(user .UserID).Handle}}{{end}}

{{define "main"}}
  {{$user := (user .UserID)}}
  <div class="user-feed-header">
    {{template "user-header" $user}}

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
    {{if .PinnedTweet.ID}}
      <div class="pinned-tweet">
        <div class="row pinned-tweet__pin-container">
          <img class="svg-icon pinned-tweet__pin-icon" src="/static/icons/pin.svg" width="24" height="24" />
          <span>Pinned</span>
        </div>
        {{template "tweet" (dict "TweetID" .PinnedTweet.ID "RetweetID" 0)}}
      </div>
    {{end}}
    {{template "timeline" .}}
  </div>
{{end}}
