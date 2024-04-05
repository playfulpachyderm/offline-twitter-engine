{{define "title"}}Search{{end}}

{{define "main"}}
  <div class="search-header">
    <div class="row row--spread">
      <div class="dummy"></div> {{/* Extra div to take up a slot in the `row` */}}
      <h1>Search results: {{.SearchText}}</h1>
      <div class="row">
        <a class="button" target="_blank" href="https://twitter.com/search?q={{.SearchText}}&src=typed_query&f=top" title="Open on twitter.com">
          <img class="svg-icon" src="/static/icons/external-link.svg" width="24" height="24" />
        </a>
        <a class="button" hx-get="?scrape" hx-target="body" hx-indicator=".search-header" title="Refresh">
          <img class="svg-icon" src="/static/icons/refresh.svg" width="24" height="24" />
        </a>
      </div>
    </div>

    <div class="tabs row">
      <a class="tabs__tab {{if (not .IsUsersSearch)}}tabs__tab--active{{end}}" href="?type=tweets">
        <span class="tabs__tab-label">Tweets</span>
      </a>
      <a class="tabs__tab {{if .IsUsersSearch}}tabs__tab--active{{end}}" href="?type=users">
        <span class="tabs__tab-label">Users</span>
      </a>
    </div>
    <div class="htmx-spinner">
      <div class="htmx-spinner__background"></div>
      <img class="svg-icon htmx-spinner__icon" src="/static/icons/spinner.svg" />
    </div>
  </div>
  {{if .IsUsersSearch}}
    {{template "list" (dict "UserIDs" .UserIDs)}}
  {{else}}
    <div class="sort-order">
      <label class="sort-order__label">order:</label>
      <select class="sort-order__dropdown" name="sort-order" hx-get="#" hx-target="body" hx-push-url="true">
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
