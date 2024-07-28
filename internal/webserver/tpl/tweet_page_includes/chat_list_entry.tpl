{{define "chat-list-entry"}}
  {{$room := $.room}}
  <div class="chat-list-entry {{if .is_active}}chat-list-entry--active-chat{{end}} {{if .is_unread}}chat-list-entry--unread{{end}}"
    hx-get="/messages/{{$room.ID}}"
    hx-push-url="true"
    hx-swap="outerHTML"
    hx-target="body"
  >
    <div class="chat-list-entry__header">
      {{template "chat-profile-image" $room}}
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
        {{ $message := (dm_message $room.LastMessageID)}}
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
