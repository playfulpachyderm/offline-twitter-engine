{{define "nav-sidebar"}}
  <div class="nav-sidebar">
    <div id="logged-in-user-info">
      <div class="quick-link" hx-get="/login" hx-trigger="click" hx-target="body" hx-push-url="true">
        {{template "author-info" active_user}}
        <img class="svg-icon" src="/static/icons/dotdotdot.svg"  />
      </div>
    </div>
    <ul class="quick-links">
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/home.svg" />
          <span>Home</span>
        </li>
      </a>
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/explore.svg" />
          <span>Explore</span>
        </li>
      </a>
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/notifications.svg" />
          <span>Notifications</span>
        </li>
      </a>
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/messages.svg" />
          <span>Messages</span>
        </li>
      </a>
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/lists.svg" />
          <span>Lists</span>
        </li>
      </a>
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/bookmarks.svg" />
          <span>Bookmarks</span>
        </li>
      </a>
      <a class="unstyled-link" href="#">
      <li class="quick-link">
          <img class="svg-icon" src="/static/icons/communities.svg" />
        <span>Communities</span>
      </li>
      </a>
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/verified.svg" />
          <span>Verified</span>
        </li>
      </a>
      {{if (not (eq (active_user).Handle "[nobody]"))}}
        <a class="unstyled-link" href="/{{(active_user).Handle}}">
          <li class="quick-link">
            <img class="svg-icon" src="/static/icons/profile.svg" />
            <span>Profile</span>
          </li>
        </a>
      {{end}}
      <a class="unstyled-link" href="#">
        <li class="quick-link">
          <img class="svg-icon" src="/static/icons/more.svg" />
          <span>More</span>
        </li>
      </a>
    </ul>
  </div>
{{end}}
