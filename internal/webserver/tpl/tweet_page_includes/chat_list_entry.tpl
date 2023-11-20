{{define "chat-list-entry"}}
  {{$room := $.room}}
  <div class="chat {{if .is_active}}active-chat{{end}}" hx-get="/messages/{{$room.ID}}" hx-push-url="true" hx-swap="outerHTML" hx-target="body">
    <div class="chat-preview-header">
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
        <div class="groupchat-profile-image-container">
          <img class="profile-image" src="{{$room.AvatarImageRemoteURL}}" />
          <div class="display-name row">{{$room.Name}}</div>
        </div>
      {{end}}
      <div class="chat-preview-timestamp .posted-at-container">
        <p class="posted-at">
          {{$room.LastMessagedAt.Time.Format "Jan 2, 2006"}}
          <br/>
          {{$room.LastMessagedAt.Time.Format "3:04 pm"}}
        </p>
      </div>
    </div>
    <p class="chat-preview">{{(index $.messages $room.LastMessageID).Text}}</p>
  </div>
{{end}}
