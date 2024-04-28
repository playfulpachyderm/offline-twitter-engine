{{define "chat-list"}}
  <div class="chat-list">
    {{range .RoomIDs}}
      {{template "chat-list-entry" (dict
          "room" (index $.Rooms .)
          "messages" $.DMTrove.Messages
          "is_active" (eq $.ActiveRoomID .)
          "is_unread" (index $.UnreadRoomIDs .)
      ) }}
    {{end}}

    {{/* Scroll the active chat into view, if there is one */}}
    {{if $.ActiveRoomID}}
      <script>
        document.querySelector(".chat-list-entry.chat-list-entry--active-chat").scrollIntoViewIfNeeded(true)
      </script>
    {{end}}
  </div>
{{end}}
