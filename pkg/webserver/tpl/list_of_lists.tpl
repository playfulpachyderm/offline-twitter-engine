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
          {{template "N-profile-images" (dict "Users" .Users "MaxDisplayUsers" $max_display_users)}}
          {{if (gt (len .Users) $max_display_users)}}
            <span class="ellipsis">...</span>
          {{end}}
        </div>
        <a class="button button--danger"
          hx-delete="/lists/{{.ID}}" hx-target="body"
          onclick="return confirm('Delete this list?  Are you sure?')"
        >Delete</a>
      </div>
    {{end}}
  </div>
{{end}}
