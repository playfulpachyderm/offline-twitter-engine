{{define "chat-list"}}
  <div class="chat-list">
    {{range .RoomIDs}}
      {{$room :=  (index $.Rooms .)}}
      <div class="chat" hx-get="/messages/{{$room.ID}}" hx-target="#chat-view" hx-swap="outerHTML" hx-push-url="true">
        {{range $room.Participants}}
          {{if (ne .UserID (active_user).ID)}}
            <!-- This is some fuckery; I have no idea why "hx-target" is needed, but otherwise it targets the #chat-view. -->
            <div class="click-eater" hx-trigger="click consume" hx-target="body">
              {{template "author-info" (user .UserID)}}
            </div>
          {{end}}
        {{end}}
        <p class="chat-preview">{{(index $.DMTrove.Messages $room.LastMessageID).Text}}</p>
      </div>
    {{end}}
  </div>
{{end}}
