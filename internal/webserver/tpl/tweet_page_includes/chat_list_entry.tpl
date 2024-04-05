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
    <p class="chat-list-entry__message-preview">{{(index $.messages $room.LastMessageID).Text}}</p>
  </div>
{{end}}
