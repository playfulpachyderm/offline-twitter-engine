{{define "list"}}
  <div class="users-list">
    {{range .UserIDs}}
      {{$user := (user .)}}
      <div class="user">
        <div class="row row--spread">
          {{template "author-info" $user}}
          {{if $.button_text}}
            <a
              href="{{$.button_url}}?user_handle={{$user.Handle}}"
              class="button button--danger"
              onclick="return confirm('{{$.button_text}} this user?  Are you sure?')"
            >
              {{$.button_text}}
            </a>
          {{end}}
        </div>
        <p class="bio">{{$user.Bio}}</p>
      </div>
    {{end}}
  </div>
{{end}}
