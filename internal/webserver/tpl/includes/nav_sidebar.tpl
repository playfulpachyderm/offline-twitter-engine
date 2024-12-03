{{define "nav-sidebar"}}
  <nav id="nav-sidebar" class="nav-sidebar" hx-trigger="load delay:3s" hx-get="/nav-sidebar-poll-updates" hx-swap="outerHTML">
    <div id="logged-in-user-info">
      <div class="button row" hx-get="/login" hx-trigger="click" hx-target="body" hx-push-url="true">
        {{template "author-info" active_user}}
        <img class="svg-icon" src="/static/icons/dotdotdot.svg" width="24" height="24" />
      </div>
    </div>
    <ul class="nav-sidebar__buttons">
      <a href="/timeline">
        <li class="button labelled-icon">
          <img class="svg-icon" src="/static/icons/home.svg" width="24" height="24" />
          <label class="nav-sidebar__button-label">Home</label>
        </li>
      </a>
      <a onclick="document.querySelector('#searchBar').focus()">
        <li class="button labelled-icon">
          <img class="svg-icon" src="/static/icons/explore.svg" width="24" height="24" />
          <label class="nav-sidebar__button-label">Explore</label>
        </li>
      </a>
      {{if (not (eq (active_user).Handle "[nobody]"))}}
        <a href="/notifications">
          <li class="nav-sidebar__notifications button labelled-icon">
            <img class="svg-icon" src="/static/icons/notifications.svg" width="24" height="24" />
            {{if .NumRegularNotifications}}
              <span class="nav-sidebar__notifications-count">{{.NumRegularNotifications}}</span>
            {{end}}
            <label class="nav-sidebar__button-label">Notifications</label>
          </li>
        </a>
        <a href="/messages">
          <li class="nav-sidebar__messages button labelled-icon">
            <img class="svg-icon" src="/static/icons/messages.svg" width="24" height="24" />
            {{if .NumMessageNotifications}}
              <span class="nav-sidebar__notifications-count">{{.NumMessageNotifications}}</span>
            {{end}}
            <label class="nav-sidebar__button-label">Messages</label>
          </li>
        </a>
      {{end}}
      <a href="/lists">
        <li class="button labelled-icon">
          <img class="svg-icon" src="/static/icons/lists.svg" width="24" height="24" />
          <label class="nav-sidebar__button-label">Lists</label>
        </li>
      </a>
      <a href="/bookmarks">
        <li class="button labelled-icon">
          <img class="svg-icon" src="/static/icons/bookmarks.svg" width="24" height="24" />
          <label class="nav-sidebar__button-label">Bookmarks</label>
        </li>
      </a>
      <a hx-get="/communities">
      <li class="button labelled-icon">
        <img class="svg-icon" src="/static/icons/communities.svg" width="24" height="24" />
        <label class="nav-sidebar__button-label">Communities</label>
      </li>
      </a>
      <a href="#">
        <li class="button labelled-icon">
          <img class="svg-icon" src="/static/icons/verified.svg" width="24" height="24" />
          <label class="nav-sidebar__button-label">Verified</label>
        </li>
      </a>
      {{if (not (eq (active_user).Handle "[nobody]"))}}
        <a href="/{{(active_user).Handle}}">
          <li class="button labelled-icon">
            <img class="svg-icon" src="/static/icons/profile.svg" width="24" height="24" />
            <label class="nav-sidebar__button-label">Profile</label>
          </li>
        </a>
      {{end}}
      <a href="https://twitter.com" target="_blank">
        <li class="button labelled-icon">
          <img class="svg-icon" src="/static/icons/internet.svg" width="24" height="24" />
          <label class="nav-sidebar__button-label">Go online</label>
        </li>
      </a>
      <a href="#">
        <li class="button labelled-icon">
          <img class="svg-icon" src="/static/icons/more.svg" width="24" height="24"/>
          <label class="nav-sidebar__button-label">More</label>
        </li>
      </a>
    </ul>
  </nav>
{{end}}
