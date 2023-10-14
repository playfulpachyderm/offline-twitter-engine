{{define "title"}}Search{{end}}

{{define "main"}}
  <div class="search-header">
    <h2>Search results: {{.SearchText}}</h2>
    <select name="sort-order" style="text-transform: capitalize;" hx-get="#" hx-target="body" hx-push-url="true">
      {{range .SortOrderOptions}}
        <option
          value="{{.}}"
          {{if (eq ($.SortOrder.String) .)}} selected {{end}}
          style="text-transform: capitalize;"
        >{{.}}</option>
      {{end}}
    </select>
  </div>
  <div class="timeline">
    {{template "timeline" .}}
  </div>
{{end}}
