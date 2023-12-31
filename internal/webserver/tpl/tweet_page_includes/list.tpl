{{define "list"}}
  <div class="users-list-container">
    {{range .}}
      {{$user := (user .)}}
      <div class="user">
        {{template "author-info" $user}}
        <p class="bio">{{$user.Bio}}</p>
      </div>
    {{end}}
  </div>
{{end}}
