{{define "list"}}
  <div class="users-list-container">
    {{range .UserIDs}}
      {{$user := (user .)}}
      <div class="user">
        <div class="row spread">
          {{template "author-info" $user}}
          {{if $.button_text}}
            <a class="unstyled-link quick-link danger" href="{{$.button_url}}?user_handle={{$user.Handle}}"onclick="return confirm('{{$.button_text}} this user?  Are you sure?')">
              {{$.button_text}}
            </a>
          {{end}}
        </div>
        <p class="bio">{{$user.Bio}}</p>
      </div>
    {{end}}
  </div>
{{end}}
