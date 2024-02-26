{{define "title"}}Lists{{end}}

{{define "main"}}
  <h1>Lists</h1>

  <button onclick="document.querySelector('#newListDialog').showModal()">New list</button>
  <dialog id="newListDialog">
    <h3>Create new list</h3>
    <form hx-post="/lists" hx-ext="json-enc" hx-target="body" hx-push-url="true">
      <label for="name">Name</label>
      <input name="name" />
      <input type="submit" value="Create" />
    </form>
    <button onclick="document.querySelector('#newListDialog').close()">Cancel</button>
  </dialog>

  <div class="users-list-previews">
    {{range .}}
      {{$max_display_users := 10}}
      <div class="users-list-preview row spread">
        <div class="list-info-container" hx-get="/lists/{{.ID}}" hx-trigger="click" hx-target="body" hx-push-url="true">
          <span class="list-name">{{.Name}}</span>
          <span class="num-users">({{(len .Users)}})</span>
          <div class="first-N-profile-images" hx-trigger="click consume">
            {{range $i, $user := .Users}}
              {{/* Only render the first 10-ish users */}}
              {{if (lt $i $max_display_users)}}
                <a class="unstyled-link" href="/{{$user.Handle}}">
                  <img
                    class="profile-image"
                    {{if $user.IsContentDownloaded}}
                      src="/content/{{$user.GetProfileImageLocalPath}}"
                    {{else}}
                      src="{{$user.ProfileImageUrl}}"
                    {{end}}
                  />
                </a>
              {{end}}
            {{end}}
            {{if (gt (len .Users) $max_display_users)}}
              <span class="ellipsis">...</span>
            {{end}}
          </div>
        </div>
        <a class="unstyled-link quick-link danger"
          hx-delete="/lists/{{.ID}}" hx-target="body"
          onclick="return confirm('Delete this list?  Are you sure?')"
        >Delete</a>
      </div>
    {{end}}
  </div>
{{end}}
