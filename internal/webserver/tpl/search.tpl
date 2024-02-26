{{define "title"}}Search{{end}}

{{define "main"}}
  <div class="search-header">
    <h1>Search results: {{.SearchText}}</h1>

    <div class="row tabs-container">
      <a
        class="tab unstyled-link {{if (not .IsUsersSearch)}}active-tab{{end}}"
        href="?type=tweets"
      >
        <span class="tab-inner">Tweets</span>
      </a>
      <a
        class="tab unstyled-link {{if .IsUsersSearch}}active-tab{{end}}"
        href="?type=users"
      >
        <span class="tab-inner">Users</span>
      </a>
    </div>
  </div>
  {{if .IsUsersSearch}}
    {{template "list" (dict "UserIDs" .UserIDs)}}
  {{else}}
    <div class="sort-order-container">
      <span class="sort-order-label">order:</span>
      <select name="sort-order" hx-get="#" hx-target="body" hx-push-url="true">
        {{range .SortOrderOptions}}
          <option
            value="{{.}}"
            style="text-transform: capitalize;"
            {{if (eq ($.SortOrder.String) .)}}
              selected
            {{end}}
          >{{.}}</option>
        {{end}}
      </select>
    </div>
    <div class="timeline">
      {{template "timeline" .}}
    </div>
  {{end}}
{{end}}
