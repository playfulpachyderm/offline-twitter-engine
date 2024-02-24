{{define "chat-list"}}
  <div class="chat-list">
    {{range .RoomIDs}}
      {{template "chat-list-entry" (dict "room" (index $.Rooms .) "messages" $.DMTrove.Messages "is_active" (eq $.ActiveRoomID .))}}
    {{end}}

    {{/* Scroll the active chat into view, if there is one */}}
    {{if $.ActiveRoomID}}
      <script>
        document.querySelector(".chat.active-chat").scrollIntoViewIfNeeded(true)
      </script>
    {{end}}
  </div>
{{end}}
