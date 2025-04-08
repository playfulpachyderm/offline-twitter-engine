{{define "main"}}
  <div class="timeline-header">
    <div class="tabs row">
      <a class="tabs__tab {{if (eq .ActiveTab "User feed")}}tabs__tab--active{{end}}" href="/timeline">
        <span class="tabs__tab-label">User feed</span>
      </a>
      <a class="tabs__tab {{if (eq .ActiveTab "Offline")}}tabs__tab--active{{end}}" href="/timeline/offline">
        <span class="tabs__tab-label">Offline timeline</span>
      </a>
    </div>
  </div>

  <div class="timeline">
    {{template "timeline" .Feed}}
  </div>
{{end}}
