{{define "title"}}Tweet{{end}}


{{define "main"}}
  {{range .ParentIDs}}
    {{template "tweet" .}}
  {{end}}
  {{template "tweet" .MainTweetID}}
  <hr />

  {{range .ReplyChains}}
    {{range .}}
      {{template "tweet" .}}
    {{end}}
    <hr />
  {{end}}
{{end}}
