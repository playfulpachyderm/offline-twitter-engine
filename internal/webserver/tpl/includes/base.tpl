{{define "base"}}
  <!doctype html>
  <html lang='en'>
    <head>
      <meta charset='utf-8'>
      <title>Offline Twitter | {{template "title" .}}</title>
      <link rel='stylesheet' href='/static/styles.css'>
      <link rel='shortcut icon' href='/static/img/favicon.ico' type='image/x-icon'>
      <link rel='stylesheet' href='/static/vendor/fonts.css'>
      <script src="/static/vendor/htmx.min.js" integrity="sha384-zUfuhFKKZCbHTY6aRR46gxiqszMk5tcHjsVFxnUo8VMus4kHGVdIYVbOYYNlKmHV" crossorigin="anonymous"></script>
    </head>
    <body>
      <div class="top-bar">
        <a onclick="window.history.back()" class="back-button quick-link">
          <img class="svg-icon" src="/static/icons/back.svg" />
        </a>
        <input class="search-bar" placeholder="Search" type="text" />
      </div>
      {{template "nav-sidebar"}}
      <main>
        {{template "main" .}}
      </main>
    </body>
  </html>
{{end}}
