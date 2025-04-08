{{define "main"}}
  <div class="list-feed-header">
    <h1>{{.List.Name}}</h1>

    <div class="tabs row">
      <a class="tabs__tab {{if (eq .ActiveTab "feed")}}tabs__tab--active{{end}}" href="/lists/{{.List.ID}}">
        <span class="tabs__tab-label">Feed</span>
      </a>
      <a class="tabs__tab {{if (eq .ActiveTab "users")}}tabs__tab--active{{end}}" href="/lists/{{.List.ID}}/users">
        <span class="tabs__tab-label">Users</span>
      </a>
    </div>
  </div>

  {{if (eq .ActiveTab "feed")}}
    <div class="timeline list-feed-timeline">
      {{template "timeline" .Feed}}
    </div>
  {{else}}
    <div class="add-users-container">
      <form action="/lists/{{.List.ID}}/add_user">
        <input type="text" name="user_handle" placeholder="@some_user_handle" style="width: 15em" />
        <input type="submit" value="Add user" />
      </form>
    </div>

    {{template "list" (dict "UserIDs" .UserIDs "button_text" "Remove" "button_url" (printf "/lists/%d/remove_user" .List.ID))}}
  {{end}}
{{end}}
