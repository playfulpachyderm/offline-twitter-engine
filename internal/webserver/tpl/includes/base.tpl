{{define "base"}}
  <!doctype html>
  <html lang='en'>
    <head>
      <meta charset='utf-8'>
      <title>Offline Twitter | {{template "title" .}}</title>
      <link rel='stylesheet' href='/static/styles.css'>
      <link rel='shortcut icon' href='/static/img/favicon.ico' type='image/x-icon'>
      <link rel='stylesheet' href='https://fonts.googleapis.com/css?family=Titillium+Web:400,700'>
    </head>
    <body>
      <div class="top-bar">
        <div class="back-button">
          <img class="svg-icon" src="/static/icons/back.svg" />
        </div>
        <input class="search-bar" placeholder="Search" type="text" />
      </div>
      {{template "nav-sidebar"}}
      <main>
        {{template "main" .}}
      </main>
    </body>
  </html>
{{end}}
