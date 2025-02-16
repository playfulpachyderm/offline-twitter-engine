{{/*
  Equivalent of an "author info", but for chats; could be an author info or a group-chat.
*/}}
{{define "chat-profile-image"}}
  {{if (eq .Type "ONE_TO_ONE")}}
    {{range .Participants}}
      {{if (ne .UserID (active_user).ID)}}
        <!-- This is some fuckery; I have no idea why "hx-target" is needed, but otherwise it targets the #chat-view. -->
        <div class="click-eater" hx-trigger="click consume" hx-target="body">
          {{template "author-info" (user .UserID)}}
        </div>
      {{end}}
    {{end}}
  {{else}}
    <div class="groupchat-info row">
      {{template "circle-profile-img-no-link" (dict "IsContentDownloaded" false "ProfileImageUrl" .AvatarImageRemoteURL)}}
      <div class="click-eater" hx-trigger="click consume" hx-target="body">
        <div class="groupchat-info__display-name">{{.Name}}</div>
      </div>
    </div>
  {{end}}
{{end}}
