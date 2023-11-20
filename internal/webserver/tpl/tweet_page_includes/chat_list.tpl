{{define "chat-list"}}
  <div class="chat-list">
    {{range .RoomIDs}}
      {{template "chat-list-entry" (dict "room" (index $.Rooms .) "messages" $.DMTrove.Messages "is_active" (eq $.ActiveRoomID .))}}
    {{end}}
  </div>
{{end}}
