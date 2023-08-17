{{define "tweet"}}
{{$main_tweet := (tweet .TweetID)}}
{{$author := (user $main_tweet.UserID)}}
<div class="tweet"
  {{if (not (eq $main_tweet.ID (focused_tweet_id)))}}
    hx-post="/tweet/{{$main_tweet.ID}}"
    hx-trigger="click target::not(.tweet-text)"
    hx-target="body"
    hx-swap="outerHTML"
    hx-push-url="true"
  {{end}}
>
  {{if (not (eq .RetweetID 0))}}
    {{$retweet := (retweet .RetweetID)}}
    {{$retweet_user := (user $retweet.RetweetedByID)}}
    <div class="retweet-info-container">
      <img class="svg-icon" src="/static/icons/retweet.svg" />
      <span class="retweeted-by-label">Retweeted by</span>
      <a class="retweeted-by-user" hx-get="/{{$retweet_user.Handle}}" hx-target="body" hx-swap="outerHTML" hx-push-url="true">
        {{$retweet_user.DisplayName}}
      </a>
    </div>
  {{end}}
  <div class="tweet-header-container">
    <div class="author-info-container" hx-trigger="click consume">
      {{template "author-info" $author}}
    </div>
    {{if $main_tweet.ReplyMentions}}
      <div class="reply-mentions-container" hx-trigger="click consume">
        <span class="replying-to-label">Replying&nbsp;to</span>
        <ul class="reply-mentions">
          {{range $main_tweet.ReplyMentions}}
            <li><a class="entity" href="/{{.}}">@{{.}}</a></li>
          {{end}}
        </ul>
      </div>
    {{end}}
    <div class="posted-at-container">
      <p class="posted-at">
        {{$main_tweet.PostedAt.Time.Format "Jan 2, 2006"}}
        <br/>
        {{$main_tweet.PostedAt.Time.Format "3:04 pm"}}
      </p>
    </div>
  </div>
  <div class="row">
    <span class="vertical-reply-line-container">
      <div class="vertical-reply-line">
      </div>
    </span>
    <span class="vertical-container-1">
      <div class="tweet-content">
        {{range (split "\n" $main_tweet.Text)}}
          <p class="tweet-text">
            {{.}}
          </p>
        {{end}}

        {{range $main_tweet.Images}}
          <img src="/content/images/{{.LocalFilename}}" style="max-width: 45%"/>
        {{end}}
        {{range $main_tweet.Videos}}
          <video controls hx-trigger="click consume">
            <source src="/content/videos/{{.LocalFilename}}">
          </video>
        {{end}}

        {{if $main_tweet.QuotedTweetID}}
          {{$quoted_tweet := (tweet $main_tweet.QuotedTweetID)}}
          {{$quoted_author := (user $quoted_tweet.UserID)}}
          <a href="/tweet/{{$quoted_tweet.ID}}">
            <div class="quoted-tweet">
              {{template "author-info" $quoted_author}}
              <div class="quoted-tweet-content">
                <a href="/tweet/{{$quoted_tweet.ID}}" class="unstyled-link tweet-text">
                  {{$quoted_tweet.Text}}
                </a>
                {{range $quoted_tweet.Images}}
                  <img src="{{.RemoteURL}}" style="max-width: 45%"/>
                {{end}}
                <p>{{$quoted_tweet.PostedAt.Time.Format "Jan 2, 2006"}}</p>
              </div>
            </div>
          </a>
        {{end}}
      </div>

      <div class="interactions-bar">
<!--         <div class="interaction-stat">
          {template "quote-tweet-icon"}
          <span>{{$main_tweet.NumQuoteTweets}}</span>
        </div> -->
        <div class="interaction-stat">
          <img class="svg-icon" src="/static/icons/reply.svg" />
          <span>{{$main_tweet.NumReplies}}</span>
        </div>
        <div class="interaction-stat">
          <img class="svg-icon" src="/static/icons/retweet.svg" />
          <span>{{$main_tweet.NumRetweets}}</span>
        </div>
        <div class="interaction-stat">
          <img class="svg-icon" src="/static/icons/like.svg" />
          <span>{{$main_tweet.NumLikes}}</span>
        </div>
        <div class="dummy"></div>
      </div>
    </span>
  </div>
</div>
{{end}}
