{{define "title"}}Notifications{{end}}

{{define "main"}}
  <div class="notifications-header">
    <div class="row row--spread">
      <div class="dummy"></div> {{/* Extra div to take up a slot in the `row` */}}
      <h1>Notifications</h1>
      <div class="row">
        <a class="button" target="_blank" href="https://twitter.com/notifications" title="Open on twitter.com">
          <img class="svg-icon" src="/static/icons/external-link.svg" width="24" height="24" />
        </a>
        <a class="button" hx-get="?scrape" hx-target="body" hx-indicator=".search-header" title="Refresh">
          <img class="svg-icon" src="/static/icons/refresh.svg" width="24" height="24" />
        </a>
      </div>
    </div>
  </div>

  <div class="timeline">
    {{template "timeline" .}}
  </div>
{{end}}
