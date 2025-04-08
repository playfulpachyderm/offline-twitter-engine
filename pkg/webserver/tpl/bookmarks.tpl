{{define "main"}}
  <div class="bookmarks-feed-header">
    <div class="row row--spread">
      <div class="dummy"></div> {{/* Extra div to take up a slot in the `row` */}}
      <h1>Bookmarks</h1>
      <div class="row">
        <a class="button" target="_blank" href="https://twitter.com/i/bookmarks" title="Open on twitter.com">
          <img class="svg-icon" src="/static/icons/external-link.svg" width="24" height="24" />
        </a>
        <a class="button" hx-get="?scrape" hx-target="body" hx-indicator=".bookmarks-feed-header" title="Refresh">
          <img class="svg-icon" src="/static/icons/refresh.svg" width="24" height="24" />
        </a>
      </div>
    </div>
    <div class="htmx-spinner">
      <div class="htmx-spinner__fullscreen-forcer">
        <div class="htmx-spinner__background"></div>
        <img class="svg-icon htmx-spinner__icon" src="/static/icons/spinner.svg" />
      </div>
    </div>
  </div>
  <div class="timeline">
    {{template "timeline" .Feed}}
  </div>
{{end}}
