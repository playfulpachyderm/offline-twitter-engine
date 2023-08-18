{{define "space"}}
  <div class="space">
    <div class="space-host row">
      {{template "author-info" (user .CreatedById)}}
      <span class="host-label">(Host)</span>
      <div class="layout-spacer"></div>
      <div class="space-date">
        {{.StartedAt.Format "Jan 2, 2006"}}<br>{{.StartedAt.Format "3:04pm"}}
      </div>
    </div>
    <h3 class="space-title">{{.Title}}</h3>
    <div class="space-info row">
      <span class="space-state">
        {{if (eq .State "Ended")}}
          <ul class="space-info-list inline-dotted-list">
            <li>{{.State}}</li>
            <li>{{(len .ParticipantIds)}} participants</li>
            <li>{{.LiveListenersCount}} tuned in</li>
            <li>Lasted {{.FormatDuration}}</li>
          </ul>
        {{else}}
          {{.State}}
        {{end}}
      </span>
    </div>
    <ul class="space-participants-list">
      {{range .ParticipantIds}}
        {{if (ne . $.CreatedById)}}
          <li>{{template "author-info" (user .)}}</li>
        {{end}}
      {{end}}
    </ul>
  </div>
{{end}}
