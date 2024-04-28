{{define "chat-list-entry"}}
  {{$room := $.room}}
  <div class="chat-list-entry {{if .is_active}}chat-list-entry--active-chat{{end}}"
    hx-get="/messages/{{$room.ID}}"
    hx-push-url="true"
    hx-swap="outerHTML"
    hx-target="body"
  >
    <div class="chat-list-entry__header">
      {{if (eq $room.Type "ONE_TO_ONE")}}
        {{range $room.Participants}}
          {{if (ne .UserID (active_user).ID)}}
            <!-- This is some fuckery; I have no idea why "hx-target" is needed, but otherwise it targets the #chat-view. -->
            <div class="click-eater" hx-trigger="click consume" hx-target="body">
              {{template "author-info" (user .UserID)}}
            </div>
          {{end}}
        {{end}}
      {{else}}
        <div class="chat-list-entry__groupchat-profile-image">
          {{template "circle-profile-img-no-link" (dict "IsContentDownloaded" false "ProfileImageUrl" $room.AvatarImageRemoteURL)}}
          <div class="click-eater" hx-trigger="click consume" hx-target="body">
            <div class="display-name row">{{$room.Name}}</div>
          </div>
        </div>
      {{end}}
      <div class="posted-at">
        <p class="posted-at__text">
          {{$room.LastMessagedAt.Time.Format "Jan 2, 2006"}}
          <br/>
          {{$room.LastMessagedAt.Time.Format "3:04 pm"}}
        </p>
      </div>
    </div>
    <div class="chat-list-entry__preview-and-unread-container row">
      <p class="chat-list-entry__message-preview">
        {{ $message := (index $.messages $room.LastMessageID)}}
        {{ $sender  := (user $message.SenderID) }}
        {{if ne $message.Text ""}}
          {{if eq $room.Type "GROUP_DM"}}
            {{ $sender.DisplayName }}:
          {{end}}
          {{$message.Text}}
        {{else if $message.EmbeddedTweetID}}
          <span class="chat-list-entry__preview-no-text">{{$sender.DisplayName}} sent a Tweet</span>
        {{else if $message.Images}}
          <span class="chat-list-entry__preview-no-text">{{$sender.DisplayName}} sent an image</span>
        {{else if $message.Videos}}
          <span class="chat-list-entry__preview-no-text">{{$sender.DisplayName}} sent a video</span>
        {{else if $message.Urls}}
          <span class="chat-list-entry__preview-no-text">{{$sender.DisplayName}} sent a link</span>
        {{end}}
      </p>
      <span class="chat-list-entry__unread-indicator"></span>
    </div>
  </div>
{{end}}
