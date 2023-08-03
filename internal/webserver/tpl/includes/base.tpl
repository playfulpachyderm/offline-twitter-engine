{{define "base"}}
  <!doctype html>
  <html lang='en'>
    <head>
      <meta charset='utf-8'>
      <title>Offline Twitter | {{template "title" .}}</title>
      <link rel='stylesheet' href='/static/styles.css'>
      <link rel='shortcut icon' href='/static/img/favicon.ico' type='image/x-icon'>
      <!-- <link rel='stylesheet' href='https://fonts.googleapis.com/css?family=Ubuntu+Mono:400,700'> -->
    </head>
    <body>
      <header>
        <h1><a href='/'>Uhhhh</a></h1>
      </header>
      <main>
        {{template "main" .}}
      </main>
    </body>
  </html>
{{end}}
