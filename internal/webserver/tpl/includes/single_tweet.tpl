{{define "tweet"}}
<div class="tweet">
  {{$main_tweet := (tweet .)}}
  {{$author := (user $main_tweet.UserID)}}

  {{template "author-info" $author}}
  <div class="tweet-content">
    <a href="/tweet/{{$main_tweet.ID}}" style="color: inherit; text-decoration: none" >{{$main_tweet.Text}}</a>

    {{range $main_tweet.Images}}
      <img src="{{.RemoteURL}}" style="max-width: 45%"/>
    {{end}}

    {{if $main_tweet.QuotedTweetID}}
      {{$quoted_tweet := (tweet $main_tweet.QuotedTweetID)}}
      {{$quoted_author := (user $quoted_tweet.UserID)}}
      <a href="/tweet/{{$quoted_tweet.ID}}">
        <div class="quoted-tweet" style="padding: 20px; outline-color: lightgray; outline-style: solid; outline-width: 1px; border-radius: 20px">
          {{template "author-info" $quoted_author}}
          <div class="quoted-tweet-content">
            <p>{{$quoted_tweet.Text}}</p>
            {{range $quoted_tweet.Images}}
              <img src="{{.RemoteURL}}" style="max-width: 45%"/>
            {{end}}
            <p>{{$quoted_tweet.PostedAt}}</p>
          </div>
        </div>
      </a>
    {{end}}

    <p>{{$main_tweet.PostedAt}}</p>
  </div>

  <div class="interactions-bar">
    <span>{{$main_tweet.NumQuoteTweets}} QTs</span>
    <span>{{$main_tweet.NumReplies}} replies</span>
    <span>{{$main_tweet.NumRetweets}} retweets</span>
    <span>{{$main_tweet.NumLikes}} likes</span>
  </div>
  <div class="interaction-buttons">

  </div>
</div>
{{end}}
