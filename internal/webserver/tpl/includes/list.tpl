{{define "list"}}
  <div class="users-list-container">
    {{range .}}
      <div class="user">
        {{template "author-info" .}}
        <p class="bio">{{.Bio}}</p>
      </div>
    {{end}}
  </div>
{{end}}
