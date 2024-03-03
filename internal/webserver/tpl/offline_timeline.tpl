{{define "title"}}Timeline{{end}}

{{define "main"}}
  <div class="timeline-header">
    <div class="row tabs-container">
      <a class="tab unstyled-link {{if (eq .ActiveTab "User feed")}}active-tab{{end}}" href="/timeline">
        <span class="tab-inner">User feed</span>
      </a>
      <a class="tab unstyled-link {{if (eq .ActiveTab "Offline")}}active-tab{{end}}" href="/timeline/offline">
        <span class="tab-inner">Offline timeline</span>
      </a>
    </div>
  </div>

  <div class="timeline">
    {{template "timeline" .Feed}}
  </div>
{{end}}
