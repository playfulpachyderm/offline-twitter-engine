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
    <button onclick="newListDialog.close()">Cancel</button>
  </dialog>

  <div class="list-of-lists">
    {{range .}}
      {{$max_display_users := 10}}
      <div class="list-preview row row--spread">
        <div class="list-preview__info-container" hx-get="/lists/{{.ID}}" hx-trigger="click" hx-target="body" hx-push-url="true">
          <span class="list-name">{{.Name}}</span>
          <span class="list-preview__num-users">({{(len .Users)}})</span>
          <div class="list-preview__first-N-profile-images" hx-trigger="click consume">
            {{range $i, $user := .Users}}
              {{/* Only render the first 10-ish users */}}
              {{if (lt $i $max_display_users)}}
                {{template "circle-profile-img" $user}}
              {{end}}
            {{end}}
            {{if (gt (len .Users) $max_display_users)}}
              <span class="ellipsis">...</span>
            {{end}}
          </div>
        </div>
        <a class="button button--danger"
          hx-delete="/lists/{{.ID}}" hx-target="body"
          onclick="return confirm('Delete this list?  Are you sure?')"
        >Delete</a>
      </div>
    {{end}}
  </div>
{{end}}
