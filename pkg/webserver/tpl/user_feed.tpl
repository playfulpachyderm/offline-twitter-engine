{{define "title"}}@{{(user .UserID).Handle}}{{end}}

{{define "main"}}
  {{$user := (user .UserID)}}
  <div class="user-feed-header">
    {{template "user-header" $user}}

    <div class="tabs row">
      <a class="tabs__tab {{if (eq .FeedType "")}}tabs__tab--active{{end}}" href="/{{$user.Handle}}">
        <span class="tabs__tab-label">Tweets and replies</span>
      </a>
      <a class="tabs__tab {{if (eq .FeedType "without_replies")}}tabs__tab--active{{end}}" href="/{{$user.Handle}}/without_replies">
        <span class="tabs__tab-label">Tweets</span>
      </a>
      <a class="tabs__tab {{if (eq .FeedType "media")}}tabs__tab--active{{end}}" href="/{{$user.Handle}}/media">
        <span class="tabs__tab-label">Media</span>
      </a>
      <a class="tabs__tab {{if (eq .FeedType "likes")}}tabs__tab--active{{end}}" href="/{{$user.Handle}}/likes">
        <span class="tabs__tab-label">Likes</span>
      </a>
    </div>
  </div>

  <div class="timeline user-feed-timeline">
    {{if .PinnedTweet.ID}}
      <div class="pinned-tweet">
        <div class="pinned-tweet__pin-container labelled-icon">
          <img class="svg-icon pinned-tweet__pin-icon" src="/static/icons/pin.svg" width="24" height="24" />
          <label>Pinned</label>
        </div>
        {{template "tweet" (dict "TweetID" .PinnedTweet.ID "RetweetID" 0 "QuoteNestingLevel" 0)}}
      </div>
    {{end}}
    {{template "timeline" .}}
  </div>
{{end}}
