{{define "nav-sidebar"}}
  <div class="nav-sidebar">
    <div id="logged-in-user-info">
      <div class="quick-link" hx-get="/login" hx-trigger="click" hx-target="body" hx-push-url="true">
        {{template "author-info" active_user}}
        <img class="svg-icon" src="/static/icons/dotdotdot.svg" width="24" height="24" />
      </div>
    </div>
    <ul class="quick-links">
      <a class="unstyled-link" href="/timeline">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/home.svg" width="24" height="24" />
          <span>Home</span>
        </li>
      </a>
      <a class="unstyled-link" onclick="document.querySelector('#search-bar').focus()">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/explore.svg" width="24" height="24" />
          <span>Explore</span>
        </li>
      </a>
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/notifications.svg" width="24" height="24" />
          <span>Notifications</span>
        </li>
      </a>
      {{if (not (eq (active_user).Handle "[nobody]"))}}
        <a class="unstyled-link" href="/messages">
          <li class="quick-link">
            <img class="svg-icon" src="/static/icons/messages.svg" width="24" height="24" />
            <span>Messages</span>
          </li>
        </a>
      {{end}}
      <a class="unstyled-link" href="/lists">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/lists.svg" width="24" height="24" />
          <span>Lists</span>
        </li>
      </a>
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/bookmarks.svg" width="24" height="24" />
          <span>Bookmarks</span>
        </li>
      </a>
      <a class="unstyled-link" href="#">
      <li class="quick-link">
          <img class="svg-icon" src="/static/icons/communities.svg" width="24" height="24" />
        <span>Communities</span>
      </li>
      </a>
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/verified.svg" width="24" height="24" />
          <span>Verified</span>
        </li>
      </a>
      {{if (not (eq (active_user).Handle "[nobody]"))}}
        <a class="unstyled-link" href="/{{(active_user).Handle}}">
          <li class="quick-link">
            <img class="svg-icon" src="/static/icons/profile.svg" width="24" height="24" />
            <span>Profile</span>
          </li>
        </a>
      {{end}}
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/more.svg" width="24" height="24"/>
          <span>More</span>
        </li>
      </a>
    </ul>
  </div>
{{end}}
