{{define "space"}}
  <div class="space">
    <div class="space__host row">
      {{template "author-info" (user .CreatedById)}}
      <span class="space__host__label">(Host)</span>
      <div class="space__layout-spacer"></div>
      <div class="space__date">
        {{.StartedAt.Format "Jan 2, 2006"}}<br>{{.StartedAt.Format "3:04pm"}}
      </div>
    </div>
    <h3 class="space__title">{{.Title}}</h3>
    <div class="space__info row">
      <span class="space-state">
        {{if (eq .State "Ended")}}
          <ul class="space__info__list inline-dotted-list">
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
    <ul class="space__participants-list">
      {{range .ParticipantIds}}
        {{if (ne . $.CreatedById)}}
          <li>{{template "author-info" (user .)}}</li>
        {{end}}
      {{end}}
    </ul>
  </div>
{{end}}
