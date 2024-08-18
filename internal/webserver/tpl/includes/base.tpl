{{define "base"}}
  <!doctype html>
  <html lang='en'>
    <head>
      <meta charset='utf-8'>
      <title>{{template "title" .}} | Offline Twitter</title>
      <link rel='stylesheet' href='/static/styles.css'>
      <link rel='shortcut icon' href='/static/twitter.ico' type='image/x-icon'>
      <link rel='stylesheet' href='/static/vendor/fonts.css'>
      <link rel="manifest" href="/static/pwa/manifest.json">
      <script src="/static/vendor/htmx.min.js" integrity="sha384-zUfuhFKKZCbHTY6aRR46gxiqszMk5tcHjsVFxnUo8VMus4kHGVdIYVbOYYNlKmHV" crossorigin="anonymous"></script>
      <script src="/static/vendor/htmx-extension-json-enc.js"></script>

      <script>
        if ('serviceWorker' in navigator) {
          navigator.serviceWorker.register('/static/pwa/service-worker.js')
            .then(function(registration) {
              console.log('Service Worker registered with scope:', registration.scope);
            }).catch(function(error) {
              console.log('Service Worker registration failed:', error);
            });
        }
      </script>
    </head>
    <body>
      <header class="row search-bar">
        <a onclick="window.history.back()" class="button search-bar__back-button">
          <img class="svg-icon" src="/static/icons/back.svg" width="24" height="24"/>
        </a>
        <form class="search-bar__form" hx-get="/search" hx-push-url="true" hx-target="body" hx-swap="innerHTML show:window:top">
          <input id="search-bar" class="search-bar__input"
            name="q"
            placeholder="Search" type="text"
            {{with (search_text)}} value="{{.}}" {{end}}
          />
        </form>
      </header>
      {{template "nav-sidebar" (global_data).Notifications}}
      <main>
        {{template "main" .}}
      </main>
      <dialog
        id="image_carousel"
        class="image-carousel"
        onmousedown="event.button == 0 && event.target==this && this.close()"
      >
        <div class="image-carousel__padding">
          <a class="button image-carousel__close-button" onclick="image_carousel.close()">X</a>
          <img class="image-carousel__active-image" src="">
        </div>
      </dialog>
      <div class="toasts" id="toasts">
        {{range (global_data).Toasts}}
          {{template "toast" .}}
        {{end}}
      </div>
    </body>
  </html>
{{end}}
