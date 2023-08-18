{{define "title"}}Tweet{{end}}


{{define "main"}}
  {{range .ParentIDs}}
    <div class="thread-parent-tweet">
      {{template "tweet" (dict "TweetID" . "RetweetID" 0 "QuoteNestingLevel" 0)}}
    </div>
  {{end}}
  <div class="focused-tweet">
    {{template "tweet" (dict "TweetID" .MainTweetID "RetweetID" 0 "QuoteNestingLevel" 0)}}
  </div>

  {{range .ReplyChains}}
    <div class="reply-chain">
      {{range .}}
        <div class="reply-tweet">
          {{template "tweet" (dict "TweetID" . "RetweetID" 0 "QuoteNestingLevel" 0)}}
        </div>
      {{end}}
    </div>
  {{end}}
{{end}}
