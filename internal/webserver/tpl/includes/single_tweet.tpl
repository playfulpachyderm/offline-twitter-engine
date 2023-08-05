{{define "tweet"}}
<div class="tweet">
  {{$main_tweet := (tweet .)}}
  {{$author := (user $main_tweet.UserID)}}

  <div class="tweet-header-container">
    {{template "author-info" $author}}
    {{if $main_tweet.ReplyMentions}}
      <div class="reply-mentions-container">
        <span class="replying-to-label">Replying to</span>
        <ul class="reply-mentions">
          {{range $main_tweet.ReplyMentions}}
            <li><a class="mention" href="/{{.}}">@{{.}}</a></li>
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
  <div class="horizontal-container-1">
    <span class="vertical-reply-line-container">
      <div class="vertical-reply-line">
      </div>
    </span>
    <span class="vertical-container-1">
      <div class="tweet-content">
        <a href="/tweet/{{$main_tweet.ID}}" class="unstyled-link tweet-text">
          {{$main_tweet.Text}}
        </a>

        {{range $main_tweet.Images}}
          <img src="/content/images/{{.LocalFilename}}" style="max-width: 45%"/>
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
