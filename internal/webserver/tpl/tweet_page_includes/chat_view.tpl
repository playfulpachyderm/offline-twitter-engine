{{define "messages-with-poller"}}
  {{range .MessageIDs}}
    {{$message := (index $.DMTrove.Messages .)}}
    {{$user := (user $message.SenderID)}}
    {{$is_us := (eq $message.SenderID (active_user).ID)}}
    <div class="dm-message-and-reacts-container {{if $is_us}} our-message {{end}}">
      <div class="dm-message-container">
        <div class="sender-profile-image-container">
          <a class="unstyled-link" href="/{{$user.Handle}}">
            {{if $user.IsContentDownloaded}}
              <img class="profile-image" src="/content/{{$user.GetProfileImageLocalPath}}" />
            {{else}}
              <img class="profile-image" src="{{$user.ProfileImageUrl}}" />
            {{end}}
          </a>
        </div>
        <div class="dm-message-content-container">
          {{if (ne $message.InReplyToID 0)}}
            <div class="replying-to-container">
              <div class="replying-to-label row">
                <img class="svg-icon" src="/static/icons/replying_to.svg" width="24" height="24" />
                <span>Replying to</span>
              </div>
              <div class="replying-to-message">
                {{(index $.DMTrove.Messages $message.InReplyToID).Text}}
              </div>
            </div>
          {{end}}
          {{if (ne $message.EmbeddedTweetID 0)}}
            <div class="tweet-preview">
              {{template "tweet" (dict
                "TweetID" $message.EmbeddedTweetID
                "RetweetID" 0
                "QuoteNestingLevel" 1)
              }}
            </div>
          {{end}}
          {{range $message.Images}}
            <img class="dm-embedded-image"
              {{if .IsDownloaded}}
                src="/content/images/{{.LocalFilename}}"
              {{else}}
                src="{{.RemoteURL}}"
              {{end}}
              width="{{.Width}}" height="{{.Height}}"
              onclick="image_carousel.querySelector('img').src = this.src; image_carousel.showModal();"
            >
          {{end}}
          {{range $message.Videos}}
            <video controls width="{{.Width}}" height="{{.Height}}"
              {{if .IsDownloaded}}
                poster="/content/video_thumbnails/{{.ThumbnailLocalPath}}"
              {{else}}
                poster="{{.ThumbnailRemoteUrl}}"
              {{end}}
            >
              {{if .IsDownloaded}}
                <source src="/content/videos/{{.LocalFilename}}">
              {{else}}
                <source src="{{.RemoteURL}}">
              {{end}}
            </video>
          {{end}}
          {{range $message.Urls}}
            <a
              class="embedded-link rounded-gray-outline unstyled-link"
              target="_blank"
              href="{{.Text}}"
              style="max-width: {{if (ne .ThumbnailWidth 0)}}{{.ThumbnailWidth}}px {{else}}fit-content {{end}}"
            >
              <img
                {{if .IsContentDownloaded}}
                  src="/content/link_preview_images/{{.ThumbnailLocalPath}}"
                {{else}}
                  src="{{.ThumbnailRemoteUrl}}"
                {{end}}
                class="embedded-link-preview"
                width="{{.ThumbnailWidth}}" height="{{.ThumbnailHeight}}"
              />
              <h3 class="embedded-link-title">{{.Title}}</h3>
              <p class="embedded-link-description">{{.Description}}</p>
              <span class="row embedded-link-domain-container">
                <img class="svg-icon" src="/static/icons/link3.svg" width="24" height="24" />
                <span class="embedded-link-domain">{{(.GetDomain)}}</span>
              </span>
            </a>
          {{end}}
          {{if $message.Text}}
            <div class="dm-message-text-container">
              {{template "text-with-entities" $message.Text}}
            </div>
          {{end}}
        </div>
      </div>
      <div class="dm-message-reactions">
        {{range $message.Reactions}}
          {{$sender := (user .SenderID)}}
          <span title="{{$sender.DisplayName}} (@{{$sender.Handle}})">{{.Emoji}}</span>
        {{end}}
      </div>
      <p class="posted-at">
        {{$message.SentAt.Time.Format "Jan 2, 2006 @ 3:04 pm"}}
      </p>
    </div>
  {{end}}

  <div id="new-messages-poller"
    hx-swap="outerHTML {{if $.ScrollBottom}}scroll:.chat-messages:bottom{{end}}"
    hx-trigger="load delay:3s"
    hx-get="/messages/{{$.ActiveRoomID}}?poll&latest_timestamp={{$.LatestPollingTimestamp}}&scroll_bottom={{if $.ScrollBottom}}1{{else}}0{{end}}"
  ></div>
{{end}}

{{define "chat-view"}}
  <div id="chat-view">
    <div class="chat-messages">
      {{if $.ActiveRoomID}}
        {{template "messages-with-poller" .}}
      {{end}}
    </div>
    {{if $.ActiveRoomID}}
      <div class="dm-composer-container">
        <form hx-post="/messages/{{$.ActiveRoomID}}/send" hx-target="#new-messages-poller" hx-swap="outerHTML scroll:.chat-messages:bottom" hx-ext="json-enc">
          {{template "dm-composer"}}
          <input id="real-input" type="hidden" name="text" value="" />
          <input type="submit" />
        </form>
      </div>
      <script>
        // Make pasting text work for HTML as well as plain text
        var editor = document.querySelector("#composer");
        editor.addEventListener("paste", function(e) {
          // cancel paste
          e.preventDefault();
          // get text representation of clipboard
          var text = (e.originalEvent || e).clipboardData.getData('text/plain');
          // insert text manually
          document.execCommand("insertHTML", false, text);
        });
      </script>
    {{end}}
  </div>

  <script>
    (function () {  // Wrap it all in an IIFE to avoid namespace pollution
      const chat_messages = document.querySelector('.chat-messages');

      // Disable auto-scroll-bottom on new message loads if the user has scrolled up
      chat_messages.addEventListener('scroll', function() {
        const _node = document.querySelector("#new-messages-poller");
        const node = _node.cloneNode()
        _node.remove(); // Removing and re-inserting the element cancels the HTMX polling, otherwise it will use the old values

        const scrollPosition = chat_messages.scrollTop;
        var bottomOfElement = chat_messages.scrollHeight - chat_messages.clientHeight;

        var [path, qs] = node.attributes["hx-get"].value.split("?")
        var params = new URLSearchParams(qs)

        if (scrollPosition === bottomOfElement) {
          // At bottom; new messages should be scrolled into view
          node.setAttribute("hx-swap", "outerHTML scroll:.chat-messages:bottom");
          params.set("scroll_bottom", "1")
          node.setAttribute("hx-get", [path, params.toString()].join("?"))
        } else {
          // User has scrolled up; disable auto-scrolling when new messages arrive
          node.setAttribute("hx-swap", "outerHTML");
          params.set("scroll_bottom", "0")
          node.setAttribute("hx-get", [path, params.toString()].join("?"))
        }

        chat_messages.appendChild(node);
        htmx.process(node); // Manually enable HTMX on the manually-added node
      });

      // Scroll to the bottom of the chat window on initial page load
      chat_messages.scrollTop = chat_messages.scrollHeight;
    })();
  </script>
{{end}}

{{define "dm-composer"}}
  <span
    id="composer"
    role="textbox"
    contenteditable
    {{if .}}
      {{/*
        This is a separate template so it can be OOB-swapped to clear the contents of the composer
        box after a successful DM send.  However, the chat-view itself also loads via HTMX call.

        To prevent the composer from being OOB'd on the initial page load (and thus never rendering),
        we guard the "hx-swap-oob" attr; so if this template is called with nothing as the arg, then
        it will be inlined normally (i.e., not OOB'd), and if the arg is something (e.g., a DMTrove),
        then it will be OOB'd, thus clearing the contents of the composer box.
      */}}
      hx-swap-oob="true"
    {{end}}
    oninput="var text = this.innerText; document.querySelector('#real-input').value = text"
    >
  </span>
{{end}}
