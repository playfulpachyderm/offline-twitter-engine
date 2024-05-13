{{define "messages"}}
  {{range .MessageIDs}}
    {{$message := (index $.DMTrove.Messages .)}}
    {{$user := (user $message.SenderID)}}
    {{$is_us := (eq $message.SenderID (active_user).ID)}}
    <div class="dm-message {{if $is_us}} our-message {{end}}" data-message-id="{{ $message.ID }}">
      <div class="dm-message__row row">
        <div class="dm-message__sender-profile-img">
          {{template "circle-profile-img" $user}}
        </div>
        <div class="dm-message__contents">
          {{if (ne $message.InReplyToID 0)}}
            <div class="dm-message__replying-to">
              <div class="dm-message__replying-to-label labelled-icon">
                <img class="svg-icon" src="/static/icons/replying_to.svg" width="24" height="24" />
                <label>Replying to</label>
              </div>
              <div class="replying-to-message"
                data-replying-to-message-id="{{ $message.InReplyToID }}"
                onclick="doReplyTo(this)"
              >
                {{(index $.DMTrove.Messages $message.InReplyToID).Text}}
              </div>
            </div>
          {{end}}
          {{if (ne $message.EmbeddedTweetID 0)}}
            <div class="dm-message__tweet-preview">
              {{template "tweet" (dict
                "TweetID" $message.EmbeddedTweetID
                "RetweetID" 0
                "QuoteNestingLevel" 1)
              }}
            </div>
          {{end}}
          {{range $message.Images}}
            <img class="dm-message__embedded-image"
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
            <video class="dm-message__embedded-video" controls width="{{.Width}}" height="{{.Height}}"
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
            {{template "embedded-link" .}}
          {{end}}
          {{if $message.Text}}
            <div class="dm-message__text-content">
              {{template "text-with-entities" $message.Text}}
            </div>
          {{end}}
        </div>
      </div>
      <div class="dm-message__reactions">
        {{range $message.Reactions}}
          {{$sender := (user .SenderID)}}
          <span title="{{$sender.DisplayName}} (@{{$sender.Handle}})">{{.Emoji}}</span>
        {{end}}
      </div>
      <div class="sent-at">
        <p class="sent-at__text">
          {{$message.SentAt.Time.Format "Jan 2, 2006 @ 3:04 pm"}}
        </p>
      </div>
    </div>
  {{end}}
{{end}}

{{define "messages-with-poller"}}
  {{template "messages" .}}

  <form id="new-messages-poller"
    hx-swap="outerHTML {{if $.ScrollBottom}}scroll:.chat-messages:bottom{{end}}"
    hx-trigger="load delay:3s"
    hx-get="/messages/{{$.ActiveRoomID}}"
  >
    <input type="hidden" name="poll">
    <input type="hidden" name="latest_timestamp" value="{{$.LatestPollingTimestamp}}">
    <input type="hidden" name="scroll_bottom" value="{{if $.ScrollBottom}}1{{else}}0{{end}}">
  </form>

  <script>
    /**
     * The poller's timestamp will be updated by HTMX, but the POST URL for /send needs updating too
     */
    (function() {
      const composer_form = document.querySelector(".dm-composer form");
      if (composer_form === null) {
        // Initial page load; composer isn't rendered yet
        return;
      }
      const [path, qs] = composer_form.attributes["hx-post"].value.split("?");
      const params = new URLSearchParams(qs);
      params.set("latest_timestamp", "{{.LatestPollingTimestamp}}");
      composer_form.setAttribute("hx-post", [path, params.toString()].join("?"));
      htmx.process(composer_form); // Manually enable HTMX on the manually-added node
    })();
  </script>
{{end}}

{{define "chat-view"}}
  <div id="chat-view">
    {{if .ActiveRoomID}}
      <div class="chat-header">
        {{ $room := (index $.Rooms $.ActiveRoomID) }}
        {{template "chat-profile-image" $room}}
        <div class="chat-header__buttons-container row">
          {{if (ne $room.Type "ONE_TO_ONE")}}
            <!-- Group chats need an "Info" button -->
            <a class="button" onclick="toggle_participants_view()" title="Show participants">
              <img class="svg-icon" src="/static/icons/info.svg" width="24" height="24" />
            </a>
          {{end}}
          <a class="button" href="https://twitter.com/messages/{{ $room.ID }}" target="_blank" title="Open on twitter.com">
            <img class="svg-icon" src="/static/icons/external-link.svg" width="24" height="24" />
          </a>
          <a class="button" hx-post="/messages/{{ $room.ID }}/mark-as-read" title="Mark as read">
            <img class="svg-icon" src="/static/icons/eye.svg" width="24" height="24" />
          </a>
          <a class="button" hx-post="/messages/{{ $room.ID }}?scrape" hx-target="#chat-view" hx-swap="outerHTML" title="Refresh">
            <img class="svg-icon" src="/static/icons/refresh.svg" width="24" height="24" />
          </a>
        </div>
      </div>
    {{end}}
    <div class="chat-messages">
      {{if .ActiveRoomID}}
        {{template "conversation-top" .}}
        {{template "messages-with-poller" .}}
      {{end}}
    </div>
    {{if .ActiveRoomID}}
      <div class="dm-composer">
        <form
          hx-post="/messages/{{.ActiveRoomID}}/send?latest_timestamp={{.LatestPollingTimestamp}}"
          hx-target="#new-messages-poller"
          hx-swap="outerHTML scroll:.chat-messages:bottom"
          hx-ext="json-enc"
          hx-on:htmx:after-request="composer.innerText = ''; realInput.value = ''; "
        >
          <span id="composer" role="textbox" contenteditable oninput="realInput.value = this.innerText"></span>
          <input id="realInput" type="hidden" name="text" value="" />
          <input type="submit" />
        </form>
      </div>
      <script>
        // Make pasting text work for HTML as well as plain text
        composer.addEventListener("paste", function(e) {
          // cancel paste
          e.preventDefault();
          // get text representation of clipboard
          var text = (e.originalEvent || e).clipboardData.getData('text/plain');
          // insert text manually
          document.execCommand("insertHTML", false, text);
        });
      </script>

      {{ $room := (index $.Rooms $.ActiveRoomID) }}
      {{if (ne $room.Type "ONE_TO_ONE")}}
        <div class="groupchat-participants-list">
          <div class="header row">
            <a onclick="toggle_participants_view()" class="button back-button">
              <img class="svg-icon" src="/static/icons/back.svg" width="24" height="24">
            </a>
            <h3>People</h3>
          </div>
          {{template "list" (dict "UserIDs" $room.GetParticipantIDs)}}
        </div>
      {{end}}
    {{end}}
  </div>

  <script>
    /**
     * When new messages are loaded in (via polling), they should be scrolled into view.
     * However, if the user has scrolled up in the conversation, they shouldn't be.
     * Also, when the conversation is opened, we should start at the bottom by default.
     */
    (function () {  // Wrap it all in an IIFE to avoid namespace pollution
      const chat_messages = document.querySelector('.chat-messages');

      // Disable auto-scroll-bottom on new message loads if the user has scrolled up
      chat_messages.addEventListener('scroll', function() {
        const _node = document.querySelector("#new-messages-poller");
        const node = _node.cloneNode(true)
        _node.remove(); // Removing and re-inserting the element cancels the HTMX polling, otherwise it will use the old values
        const scroll_bottom_input = node.querySelector("input[name='scroll_bottom']")

        const scrollPosition = chat_messages.scrollTop;
        var bottomOfElement = chat_messages.scrollHeight - chat_messages.clientHeight;

        if (scrollPosition === bottomOfElement) {
          // At bottom; new messages should be scrolled into view
          node.setAttribute("hx-swap", "outerHTML scroll:.chat-messages:bottom");
          scroll_bottom_input.value = 1;
        } else {
          // User has scrolled up; disable auto-scrolling when new messages arrive
          node.setAttribute("hx-swap", "outerHTML");
          scroll_bottom_input.value = 0;
        }

        chat_messages.appendChild(node);
        htmx.process(node); // Manually enable HTMX on the manually-added node
      });

      // Scroll to the bottom of the chat window on initial page load
      chat_messages.scrollTop = chat_messages.scrollHeight;
    })();

    /**
     * Define callback on-click handler for 'replying-to' previews; they should scroll the replied-to
     * message into view, if possible.
     */
    function doReplyTo(replying_to_box) {
      const replied_to_id = replying_to_box.getAttribute("data-replying-to-message-id");
      const replied_to_message = document.querySelector('[data-message-id="' + replied_to_id + '"]');

      replied_to_message.scrollIntoView({behavior: "smooth", block: "center"});
      replied_to_message.classList.add("highlighted");
      setTimeout(function() {
        replied_to_message.classList.remove("highlighted");
      }, 1000);
    }

    /**
     * Show or hide the Participants view
     */
    function toggle_participants_view() {
      const panel = document.querySelector(".groupchat-participants-list");
      if (panel.classList.contains("unhidden")) {
        panel.classList.remove("unhidden");
      } else {
        panel.classList.add("unhidden");
      }
    }
  </script>
{{end}}

{{define "conversation-top"}}
  <div class="show-more conversation-top">
    {{if .Cursor.CursorPosition.IsEnd}}
      <label class="show-more__eof-label">Beginning of conversation</label>
    {{else}}
      <a class="show-more__button button"
        hx-get="?cursor={{.Cursor.CursorValue}}" {{/* TODO: this might require a `cursor_to_query_params` if the same view is used for searching */}}
        hx-target=".conversation-top"
        hx-swap="outerHTML"
      >Show more</a>
    {{end}}
  </div>
{{end}}

{{/* convenience template for htmx requests */}}
{{define "messages-top"}}
  {{template "conversation-top" .}}
  {{template "messages" .}}
  <script>
    /**
     * Scroll the last message into view
     */
    (function() {
      const last_message = document.querySelector(
        '[data-message-id="{{ index .MessageIDs (sub (len .MessageIDs) 1)}}"]'
      );
      last_message.scrollIntoView({behavior: "instant", block: "start"});
      last_message.classList.add("highlighted");
      setTimeout(function() {
        last_message.classList.remove("highlighted");
      }, 1000);
    })();
  </script>
{{end}}
