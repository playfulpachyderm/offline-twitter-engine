{{define "title"}}{{.Title}}{{end}}

{{define "main"}}
  {{ $user := (user .HeaderUserID)}}
  {{template "user-header" $user}}

  <div class="tabs row" hx-boost="true">
    <a class="tabs__tab {{if (eq .Title "Followers")}}tabs__tab--active{{end}}" href="/{{$user.Handle}}/followers">
      <span class="tabs__tab-label">Followers</span>
    </a>
    <a class="tabs__tab {{if (eq .Title "Followers you know")}}tabs__tab--active{{end}}" href="/{{$user.Handle}}/followers_you_know">
      <span class="tabs__tab-label">Followers you know</span>
    </a>
    <a class="tabs__tab {{if (eq .Title "Followees")}}tabs__tab--active{{end}}" href="/{{$user.Handle}}/followees">
      <span class="tabs__tab-label">Followees</span>
    </a>
    <a class="tabs__tab {{if (eq .Title "Followees you know")}}tabs__tab--active{{end}}" href="/{{$user.Handle}}/followees_you_know">
      <span class="tabs__tab-label">Followees you know</span>
    </a>
  </div>
  {{template "list" (dict "UserIDs" .UserIDs)}}
{{end}}
