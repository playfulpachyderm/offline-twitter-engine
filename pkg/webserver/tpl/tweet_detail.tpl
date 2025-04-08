{{define "main"}}
  <div class="tweet-detail">
    {{range .ParentIDs}}
      <div class="thread-parent-tweet">
        {{template "tweet" (dict "TweetID" . "RetweetID" 0 "QuoteNestingLevel" 0)}}
      </div>
    {{end}}

    <div id="focused-tweet" class="focused-tweet">
      {{template "tweet" (dict "TweetID" .MainTweetID "RetweetID" 0 "QuoteNestingLevel" 0)}}
    </div>

    {{if (len .ThreadIDs)}}
      <div class="reply-chain">
        {{range .ThreadIDs}}
          <div class="reply-tweet">
            {{template "tweet" (dict "TweetID" . "RetweetID" 0 "QuoteNestingLevel" 0)}}
          </div>
        {{end}}
      </div>
    {{end}}

    {{range .ReplyChains}}
      <div class="reply-chain">
        {{range .}}
          <div class="reply-tweet">
            {{template "tweet" (dict "TweetID" . "RetweetID" 0 "QuoteNestingLevel" 0)}}
          </div>
        {{end}}
      </div>
    {{end}}
  </div>
{{end}}
