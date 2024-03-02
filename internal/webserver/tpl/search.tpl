{{define "title"}}Search{{end}}

{{define "main"}}
  <div class="search-header">
    <div class="row spread">
      <div class="dummy"></div> {{/* Extra div to take up a slot in the `row` */}}
      <h1>Search results: {{.SearchText}}</h1>
      <div class="user-feed-buttons-container">
        <a class="unstyled-link quick-link" target="_blank" href="https://twitter.com/search?q={{.SearchText}}&src=typed_query&f=top" title="Open on twitter.com">
          <img class="svg-icon" src="/static/icons/external-link.svg" width="24" height="24" />
        </a>
        <a class="unstyled-link quick-link" hx-get="?scrape" hx-target="body" hx-indicator=".search-header" title="Refresh">
          <img class="svg-icon" src="/static/icons/refresh.svg" width="24" height="24" />
        </a>
      </div>
    </div>

    <div class="row tabs-container">
      <a class="tab unstyled-link {{if (not .IsUsersSearch)}}active-tab{{end}}" href="?type=tweets">
        <span class="tab-inner">Tweets</span>
      </a>
      <a class="tab unstyled-link {{if .IsUsersSearch}}active-tab{{end}}" href="?type=users">
        <span class="tab-inner">Users</span>
      </a>
    </div>
    <div class="htmx-spinner-container">
      <div class="htmx-spinner-background"></div>
      <img class="svg-icon htmx-spinner" src="/static/icons/spinner.svg" />
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
