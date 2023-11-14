{{define "chat-view"}}
  <div id="chat-view">
    {{range .MessageIDs}}
      {{$message := (index $.DMTrove.Messages .)}}
      {{$user := (user $message.SenderID)}}
      {{$is_us := (eq $message.SenderID (active_user).ID)}}
      <div class="dm-message-and-reacts-container {{if $is_us}} our-message {{end}}">
        <div class="dm-message-container">
          <div class="sender-profile-image-container">
            <a class="unstyled-link" href="/{{$user.Handle}}">
              <img class="profile-image" src="/content/{{$user.GetProfileImageLocalPath}}" />
            </a>
          </div>
          <p class="dm-message-text">{{$message.Text}}</p>
        </div>
        <div class="dm-message-reactions">
          {{range $message.Reactions}}
            {{$sender := (user .SenderID)}}
            <span title="{{$sender.DisplayName}} (@{{$sender.Handle}})">{{.Emoji}}</span>
          {{end}}
        </div>
      </div>
    {{end}}
  </div>
{{end}}
