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
    <div class="retweet-info-container" hx-trigger="click consume">
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
        <ul class="reply-mentions inline-dotted-list">
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
        {{if (ne $main_tweet.TombstoneType "")}}
          <div class="tombstone">
            {{(get_tombstone_text $main_tweet)}}
          </div>
        {{end}}
        {{template "text-with-entities" $main_tweet.Text}}
        {{range $main_tweet.Images}}
          <img class="tweet-image"
            {{if .IsDownloaded}}
              src="/content/images/{{.LocalFilename}}"
            {{else}}
              src="{{.RemoteURL}}"
            {{end}}
            width="{{.Width}}" height="{{.Height}}"
            {{if (gt (len $main_tweet.Images) 1)}}
              style="max-width: 45%"
            {{end}}
          />
        {{end}}
        {{range $main_tweet.Videos}}
          <video controls hx-trigger="click consume" width="{{.Width}}" height="{{.Height}}"
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
        {{end}}
        {{range $main_tweet.Urls}}
          <div class="click-eater" hx-trigger="click consume">
            <a
              class="embedded-link rounded-gray-outline unstyled-link"
              target="_blank"
              href="{{.Text}}"
              style="max-width: {{if (ne .ThumbnailWidth 0)}}{{.ThumbnailWidth}}px {{else}}fit-content {{end}}"
            >
              <img
                {{if .IsContentDownloaded}}
                  src="/content/link_preview_images/{{.ThumbnailLocalPath}}"
                {{else}}
                  src="{{.ThumbnailRemoteUrl}}"
                {{end}}
                class="embedded-link-preview"
                width="{{.ThumbnailWidth}}" height="{{.ThumbnailHeight}}"
              />
              <h3 class="embedded-link-title">{{.Title}}</h3>
              <p class="embedded-link-description">{{.Description}}</p>
              <span class="row embedded-link-domain-container">
                <img class="svg-icon" src="/static/icons/link3.svg" />
                <span class="embedded-link-domain">{{(.GetDomain)}}</span>
              </span>
            </a>
          </div>
        {{end}}
        {{range $main_tweet.Polls}}
          {{template "poll" .}}
        {{end}}

        {{if (and $main_tweet.QuotedTweetID (lt .QuoteNestingLevel 1))}}
          <div class="quoted-tweet rounded-gray-outline" hx-trigger="click consume">
            {{template "tweet" (dict "TweetID" $main_tweet.QuotedTweetID "RetweetID" 0 "QuoteNestingLevel" (add .QuoteNestingLevel 1))}}
          </div>
        {{end}}
        {{if $main_tweet.SpaceID}}
          {{template "space" (space $main_tweet.SpaceID)}}
        {{end}}
      </div>

      <div class="interactions-bar row">
        <div class="interaction-stat">
          <img class="svg-icon" src="/static/icons/quote.svg" />
          <span>{{$main_tweet.NumQuoteTweets}}</span>
        </div>
        <div class="interaction-stat">
          <img class="svg-icon" src="/static/icons/reply.svg" />
          <span>{{$main_tweet.NumReplies}}</span>
        </div>
        <div class="interaction-stat">
          <img class="svg-icon" src="/static/icons/retweet.svg" />
          <span>{{$main_tweet.NumRetweets}}</span>
        </div>
        {{template "likes-count" $main_tweet}}
        <div class="dummy"></div>
        <div class="tweet-buttons-container" hx-trigger="click consume">
          <a class="unstyled-link quick-link" target="_blank" href="https://twitter.com/{{$author.Handle}}/status/{{$main_tweet.ID}}" title="Open on twitter.com">
            <img class="svg-icon" src="/static/icons/external-link.svg" />
          </a>
          <a class="unstyled-link quick-link" hx-get="/tweet/{{$main_tweet.ID}}?scrape" hx-target="body" title="Refresh">
            <img class="svg-icon" src="/static/icons/refresh.svg" />
          </a>
        </div>
      </div>
    </span>
  </div>
</div>
{{end}}
