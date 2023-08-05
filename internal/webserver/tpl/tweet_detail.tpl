{{define "title"}}Tweet{{end}}


{{define "main"}}
  {{range .ParentIDs}}
    <div class="thread-parent-tweet">
      {{template "tweet" .}}
    </div>
  {{end}}
  <div class="focused-tweet">
    {{template "tweet" .MainTweetID}}
  </div>

  {{range .ReplyChains}}
    <div class="reply-chain">
      {{range .}}
        <div class="reply-tweet">
          {{template "tweet" .}}
        </div>
      {{end}}
    </div>
  {{end}}
{{end}}
