self.addEventListener('install', function(event) {
  // Skip caching step during installation
});

self.addEventListener('activate', function(event) {
  // Do nothing special during activation
});

self.addEventListener('fetch', function(event) {
  // Bypass the service worker for network requests
  event.respondWith(fetch(event.request));
});
