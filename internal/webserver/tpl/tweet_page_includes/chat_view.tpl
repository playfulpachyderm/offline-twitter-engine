{{define "message"}}
  {{$user := (user .SenderID)}}
  {{$is_us := (eq .SenderID (active_user).ID)}}
  <div class="dm-message {{if $is_us}} our-message {{end}}" data-message-id="{{ .ID }}" hx-ext="json-enc" hx-swap="outerHTML">
    <div class="dm-message__row row">
      <div class="dm-message__sender-profile-img">
        {{template "circle-profile-img" $user}}
      </div>
      <div class="dm-message__contents">
        {{if (ne .InReplyToID 0)}}
          <div class="dm-replying-to">
            <div class="dm-replying-to__label labelled-icon">
              <img class="svg-icon" src="/static/icons/replying_to.svg" width="24" height="24" />
              <label>Replying to</label>
              <span class="dm-replying-to__username"
                data-replying-to-message-id="{{ .InReplyToID }}"
                onclick="handleReplyingToClicked(this)"
              >
                {{ (user (dm_message .InReplyToID).SenderID).DisplayName }}
              </span>
            </div>
            <div class="dm-replying-to__preview-text"
              data-replying-to-message-id="{{ .InReplyToID }}"
              onclick="handleReplyingToClicked(this)"
            >
              {{(dm_message .InReplyToID).Text}}
            </div>
          </div>
        {{end}}
        {{if (ne .EmbeddedTweetID 0)}}
          <div class="dm-message__tweet-preview">
            {{template "tweet" (dict
              "TweetID" .EmbeddedTweetID
              "RetweetID" 0
              "QuoteNestingLevel" 1)
            }}
          </div>
        {{end}}
        {{range .Images}}
          {{/* DM images can only be loaded with an authenticated session.  So sending the RemoteUrl
                in an <img> tag is not useful; it will just HTTP 401 */}}
          <img class="dm-message__embedded-image"
            src="/content/images/{{.LocalFilename}}"
            width="{{.Width}}" height="{{.Height}}"
            onclick="image_carousel.querySelector('img').src = this.src; image_carousel.showModal();"
          >
        {{end}}
        {{range .Videos}}
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
        {{range .Urls}}
          {{template "embedded-link" .}}
        {{end}}
        {{if .Text}}
          <div class="dm-message__text-content">
            {{template "text-with-entities" .Text}}
          </div>
        {{end}}
      </div>
      <div class="dm-message__button-container">
        <div class="row">
          <div class="dm-message__emoji-button button" onclick="
            show_emoji_picker(function(emoji_info) {
              htmx.ajax('POST', '/messages/{{$.DMChatRoomID}}/reacc', {values: {
                message_id: '{{$.ID}}',
                reacc: emoji_info.unicode,
              }, source: '[data-message-id=\'{{$.ID}}\']'});
            });
          ">
            <img class="svg-icon" src="/static/icons/emoji-react.svg" width="24" height="24"/>
          </div>
          <div class="dm-message__emoji-button button" onclick="handle_reply_clicked(this.closest('.dm-message').getAttribute('data-message-id'))">
            <img class="svg-icon" src="/static/icons/replying_to.svg" width="24" height="24"/>
          </div>
        </div>
      </div>
    </div>
    <div class="dm-message__reactions">
      {{range .Reactions}}
        {{$sender := (user .SenderID)}}
        <span
          class="dm-message__reacc {{if (eq $sender.ID (active_user).ID)}} dm-message__reacc--ours{{end}}"
          title="{{$sender.DisplayName}} (@{{$sender.Handle}})"
        >{{.Emoji}}</span>
      {{end}}
    </div>
    <div class="sent-at">
      <p class="sent-at__text">
        {{.SentAt.Time.Format "Jan 2, 2006 @ 3:04 pm"}}
      </p>
    </div>
  </div>
{{end}}


{{define "messages"}}
  {{range .MessageIDs}}
    {{template "message" (dm_message .)}}
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
          <a class="button" hx-post="/messages/{{ $room.ID }}?scrape" hx-target="#chat-view" hx-swap="outerHTML" title="Refresh" hx-indicator=".chat-messages">
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
      <div class="htmx-spinner">
        <div class="htmx-spinner__background"></div>
        <img class="svg-icon htmx-spinner__icon" src="/static/icons/spinner.svg" />
      </div>
    </div>
    {{if .ActiveRoomID}}
      <div class="dm-composer">
        {{/* TODO: replying-to CSS re-use is a mess */}}
        <div id="composerReplyingTo" class="dm-composer__replying-to-container">
          <div class="dm-replying-to row row--spread">
            <div>
              <div class="dm-replying-to__label labelled-icon">
                <img class="svg-icon" src="/static/icons/replying_to.svg" width="24" height="24" />
                <label>Replying to</label>
              </div>
              <div class="dm-replying-to__preview-text"
                data-replying-to-message-id=""
                onclick="handleReplyingToClicked(this)"
              >
                Lorem ipsum dolor sit amet
              </div>
            </div>
            <img class="svg-icon button" src="/static/icons/close.svg" width="24" height="24" onclick="cancel_reply()">
          </div>
        </div>

        <form
          hx-post="/messages/{{.ActiveRoomID}}/send?latest_timestamp={{.LatestPollingTimestamp}}"
          hx-target="#new-messages-poller"
          hx-swap="outerHTML scroll:.chat-messages:bottom"
          hx-ext="json-enc"
          hx-on:htmx:after-request="composer.innerText = ''; realInput.value = ''; cancel_reply();"
        >
          <img class="svg-icon button" src="/static/icons/emoji-insert.svg" width="24" height="24" onclick="
            var carat = composer.innerText.length;
            if (composer.contains(window.getSelection().anchorNode)) {
              carat = window.getSelection().anchorOffset;
            }
            show_emoji_picker(function(emoji) {
              composer.innerText = composer.innerText.substring(0, carat) + emoji.unicode + composer.innerText.substring(carat);
              composer.oninput(); // force-update the `realInput`
              let range = document.createRange();
              range.setStart(composer.childNodes[0], carat+emoji.unicode.length);
              range.setEnd(composer.childNodes[0], carat+emoji.unicode.length);
              let selection = window.getSelection();
              selection.removeAllRanges();
              selection.addRange(range);
              composer.focus();
            });
          "/>
          <span id="composer" role="textbox" contenteditable oninput="realInput.value = this.innerText"></span>
          <input id="realInput" type="hidden" name="text" value="" />
          <input id="inReplyToInput" type="hidden" name="in_reply_to_id" value="" />
          <input type="submit" />
        </form>
      </div>
      <script>
        /**
         *  Make pasting text work for HTML as well as plain text
         */
        composer.addEventListener("paste", function(e) {
          // cancel paste
          e.preventDefault();
          // get text representation of clipboard
          var text = (e.originalEvent || e).clipboardData.getData('text/plain');
          // insert text manually
          document.execCommand("insertHTML", false, text);
        });

        /**
         * Handle "Reply" button
         */
        function handle_reply_clicked(reply_to_message_id) {
          composerReplyingTo.classList.add("unhidden");
          inReplyToInput.value = reply_to_message_id;
          console.log(reply_to_message_id);
          const message_text_element = document.querySelector('[data-message-id="' + reply_to_message_id + '"] .dm-message__text-content');
          const preview_text = message_text_element.innerText;
          const preview_element = composerReplyingTo.querySelector('.dm-replying-to__preview-text');
          preview_element.innerText = preview_text;
          preview_element.setAttribute("data-replying-to-message-id", reply_to_message_id);
        }
        /**
         * Handle cancel-reply button
         */
        function cancel_reply() {
          composerReplyingTo.classList.remove("unhidden");
          inReplyToInput.value = "";
        }
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

  <dialog id="emoji_popup" onmousedown="event.button == 0 && event.target==this && close_emoji_picker()"></dialog>

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
     *
     * Takes an element holding the data instead of the data itself, because the composer box and
     * actual messages share this function; the composer box's "data-replying-to-message-id" can
     * change.
     */
    function handleReplyingToClicked(replying_to_box) {
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

    /**
     * When the emoji-react button is clicked, show the emoji picker
     */
    function show_emoji_picker(emoji_callback) {
      const PickerElement = customElements.get("emoji-picker"); // Not copied into the namespace by default
      const picker = new PickerElement();
      picker.addEventListener('emoji-click', function(emoji_event) {
        close_emoji_picker();
        emoji_callback(emoji_event.detail)
      });
      emoji_popup.appendChild(picker);
      emoji_popup.showModal();
    }
    /**
     * Callback function to close the emoji picker
     */
    function close_emoji_picker() {
      emoji_popup.close();
      emoji_popup.innerHTML = ""; // remove the picker
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
