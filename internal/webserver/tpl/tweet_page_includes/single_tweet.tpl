{{define "tweet"}}
{{$main_tweet := (tweet .TweetID)}}
{{$author := (user $main_tweet.UserID)}}
<div class="tweet"
  {{if (not (eq $main_tweet.ID (focused_tweet_id)))}}
    hx-get="/tweet/{{$main_tweet.ID}}"
    hx-trigger="click"
    hx-target="body"
    hx-swap="outerHTML show:#focused-tweet:top"
    hx-push-url="true"
  {{end}}
>
  {{if (not (eq .RetweetID 0))}}
    {{$retweet := (retweet .RetweetID)}}
    {{$retweet_user := (user $retweet.RetweetedByID)}}
    <div class="retweet-info" hx-trigger="click consume">
      <img class="svg-icon" src="/static/icons/retweet.svg" width="24" height="24" />
      <span class="retweet-info__retweeted-by-label">Retweeted by</span>
      <a
        class="retweet-info__retweeted-by-user"
        hx-get="/{{$retweet_user.Handle}}"
        hx-target="body"
        hx-swap="outerHTML"
        hx-push-url="true"
      >
        {{$retweet_user.DisplayName}}
      </a>
    </div>
  {{end}}
  <div class="tweet__header-container">
    <div class="author-info-container" hx-trigger="click consume">
      {{template "author-info" $author}}
    </div>
    {{if $main_tweet.ReplyMentions}}
      <div class="reply-mentions" hx-trigger="click consume">
        <span class="reply-mentions__dm-message__replying-to-label">Replying&nbsp;to</span>
        <ul class="reply-mentions__list inline-dotted-list">
          {{range $main_tweet.ReplyMentions}}
            <li><a class="entity" href="/{{.}}">@{{.}}</a></li>
          {{end}}
        </ul>
      </div>
    {{end}}
    <div class="posted-at">
      <p class="posted-at__text">
        {{$main_tweet.PostedAt.Time.Format "Jan 2, 2006"}}
        <br/>
        {{$main_tweet.PostedAt.Time.Format "3:04 pm"}}
      </p>
    </div>
  </div>
  <div class="row">
    <span class="string-box">
      <div class="string">
      </div>
    </span>
    <span class="tweet__vertical-container">
      <div class="tweet-content">
        {{if (ne $main_tweet.TombstoneType "")}}
          <div class="tombstone">
            {{(get_tombstone_text $main_tweet)}}
          </div>
        {{end}}
        {{template "text-with-entities" $main_tweet.Text}}
        {{range $main_tweet.Images}}
          <img class="tweet__embedded-image"
            {{if .IsDownloaded}}
              src="/content/images/{{.LocalFilename}}"
            {{else}}
              src="{{.RemoteURL}}"
            {{end}}
            width="{{.Width}}" height="{{.Height}}"
            {{if (gt (len $main_tweet.Images) 1)}}
              style="max-width: 45%"
            {{end}}
            hx-trigger="click consume"
            onclick="image_carousel.querySelector('img').src = this.src; image_carousel.showModal();"
          >
        {{end}}
        {{range $main_tweet.Videos}}
          <div class="video">
            {{if .IsGif}}
              <div class="video__gif-controls labelled-icon">
                <img class="svg-icon" src="/static/icons/play.svg" width="24" height="24" />
                <label class="video__gif-label">GIF</label>
              </div>
              <script>
                function gif_on_click(video) {
                  if (video.paused) {
                    video.play();
                    video.parentElement.querySelector(".svg-icon").src = "/static/icons/pause.svg";
                  } else {
                    video.pause();
                    video.parentElement.querySelector(".svg-icon").src = "/static/icons/play.svg";
                  }
                }
              </script>
            {{end}}
            <video hx-trigger="click consume" width="{{.Width}}" height="{{.Height}}"
              {{if .IsGif}}
                loop muted playsinline onclick="gif_on_click(this)" class="gif"
              {{else}}
                controls class="video"
              {{end}}
              {{if .IsDownloaded}}
                poster="/content/video_thumbnails/{{.ThumbnailLocalPath}}"
              {{else}}
                poster="{{.ThumbnailRemoteUrl}}"
              {{end}}
            >
              {{if .IsDownloaded}}
                <source src="/content/videos/{{.LocalFilename}}">
              {{else}}
                <source src="{{.RemoteURL}}">
              {{end}}
            </video>
          </div>
        {{end}}
        {{range $main_tweet.Urls}}
          <div class="click-eater" hx-trigger="click consume">
            {{template "embedded-link" .}}
          </div>
        {{end}}
        {{range $main_tweet.Polls}}
          {{template "poll" .}}
        {{end}}

        {{if (and $main_tweet.QuotedTweetID (lt .QuoteNestingLevel 1))}}
          <div class="tweet__quoted-tweet rounded-gray-outline" hx-trigger="click consume">
            {{template "tweet" (dict
                "TweetID" $main_tweet.QuotedTweetID
                "RetweetID" 0
                "QuoteNestingLevel" (add .QuoteNestingLevel 1)
            ) }}
          </div>
        {{end}}
        {{if $main_tweet.SpaceID}}
          {{template "space" (space $main_tweet.SpaceID)}}
        {{end}}
      </div>

      <div class="interactions row">
        <div class="interactions__stat">
          <img class="svg-icon" src="/static/icons/quote.svg" width="24" height="24" />
          <span>{{$main_tweet.NumQuoteTweets}}</span>
        </div>
        <div class="interactions__stat">
          <img class="svg-icon" src="/static/icons/reply.svg" width="24" height="24" />
          <span>{{$main_tweet.NumReplies}}</span>
        </div>
        <div class="interactions__stat">
          <img class="svg-icon" src="/static/icons/retweet.svg" width="24" height="24" />
          <span>{{$main_tweet.NumRetweets}}</span>
        </div>
        {{template "likes-count" $main_tweet}}
        <div class="interactions__dummy"></div>
        <div class="row" hx-trigger="click consume">
          <a class="button" title="Copy link" onclick="navigator.clipboard.writeText('https://twitter.com/{{ $author.Handle }}/status/{{ $main_tweet.ID }}')">
            <img class="svg-icon" src="/static/icons/copy.svg" width="24" height="24" />
          </a>
          <a
            class="button"
            target="_blank"
            href="https://twitter.com/{{$author.Handle}}/status/{{$main_tweet.ID}}"
            title="Open on twitter.com"
          >
            <img class="svg-icon" src="/static/icons/external-link.svg" width="24" height="24" />
          </a>
          <a
            class="button"
            hx-get="/tweet/{{$main_tweet.ID}}?scrape"
            hx-target="body"
            hx-indicator="closest .tweet"
            title="Refresh"
          >
            <img class="svg-icon" src="/static/icons/refresh.svg" width="24" height="24" />
          </a>
        </div>
      </div>
    </span>
  </div>
  <div class="htmx-spinner">
    <div class="htmx-spinner__background"></div>
    <img class="svg-icon htmx-spinner__icon" src="/static/icons/spinner.svg" />
  </div>
</div>
{{end}}
