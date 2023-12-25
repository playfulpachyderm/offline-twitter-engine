{{define "messages-with-poller"}}
  {{range .MessageIDs}}
    {{$message := (index $.DMTrove.Messages .)}}
    {{$user := (user $message.SenderID)}}
    {{$is_us := (eq $message.SenderID (active_user).ID)}}
    <div class="dm-message-and-reacts-container {{if $is_us}} our-message {{end}}">
      <div class="dm-message-container">
        <div class="sender-profile-image-container">
          <a class="unstyled-link" href="/{{$user.Handle}}">
            {{if $user.IsContentDownloaded}}
              <img class="profile-image" src="/content/{{$user.GetProfileImageLocalPath}}" />
            {{else}}
              <img class="profile-image" src="{{$user.ProfileImageUrl}}" />
            {{end}}
          </a>
        </div>
        <div class="dm-message-content-container">
          {{if (ne $message.InReplyToID 0)}}
            <div class="replying-to-container">
              <div class="replying-to-label row">
                <img class="svg-icon" src="/static/icons/replying_to.svg" />
                <span>Replying to</span>
              </div>
              <div class="replying-to-message">
                {{(index $.DMTrove.Messages $message.InReplyToID).Text}}
              </div>
            </div>
          {{end}}
          {{if (ne $message.EmbeddedTweetID 0)}}
            <div class="tweet-preview">
              {{template "tweet" (dict
                "TweetID" $message.EmbeddedTweetID
                "RetweetID" 0
                "QuoteNestingLevel" 1)
              }}
            </div>
          {{end}}
          <div class="dm-message-text-container">
            {{template "text-with-entities" $message.Text}}
          </div>
        </div>
      </div>
      <div class="dm-message-reactions">
        {{range $message.Reactions}}
          {{$sender := (user .SenderID)}}
          <span title="{{$sender.DisplayName}} (@{{$sender.Handle}})">{{.Emoji}}</span>
        {{end}}
      </div>
      <p class="posted-at">
        {{$message.SentAt.Time.Format "Jan 2, 2006 @ 3:04 pm"}}
      </p>
    </div>
  {{end}}

  <div id="new-messages-poller"
    hx-swap="outerHTML"
    hx-trigger="load delay:7s"
    hx-get="/messages/{{$.ActiveRoomID}}?poll&latest_timestamp={{$.LatestPollingTimestamp}}"
  ></div>
{{end}}

{{define "chat-view"}}
  <div id="chat-view">
    <div class="chat-messages">
      {{if $.ActiveRoomID}}
        {{template "messages-with-poller" .}}
      {{end}}
    </div>
    {{if $.ActiveRoomID}}
      <div class="dm-composer-container">
        <form hx-post="/messages/{{$.ActiveRoomID}}/send" hx-target="body" hx-ext="json-enc">
          <span
            class="composer"
            role="textbox"
            contenteditable
            oninput="var text = this.innerText; document.querySelector('#real-input').value = text"
            >
          </span>
          <input id="real-input" type="hidden" name="text" value="" />
          <input type="submit" />
        </form>
      </div>
      <script>
        // Make pasting text work for HTML as well as plain text
        var editor = document.querySelector("span.composer");
        editor.addEventListener("paste", function(e) {
          // cancel paste
          e.preventDefault();
          // get text representation of clipboard
          var text = (e.originalEvent || e).clipboardData.getData('text/plain');
          // insert text manually
          document.execCommand("insertHTML", false, text);
        });
      </script>
    {{end}}
  </div>
{{end}}
