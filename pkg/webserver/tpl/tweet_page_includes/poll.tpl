{{define "poll-choice"}}
  <div class="row poll__choice">
    <div class="poll__choice-fill-bar {{if (.poll.IsWinner .votes)}}poll__choice-fill-bar--winner{{end}}" style="width: {{printf "%.1f" (.poll.VotePercentage .votes)}}%;"></div>
    <div class="poll__choice-info row">
      <span class="poll__choice-label">{{.label}}</span>
      <span class="poll__choice-votes">{{.votes}} ({{printf "%.1f" (.poll.VotePercentage .votes)}}%)</span>
    </div>
  </div>

{{end}}


{{define "poll"}}
  <div class="poll rounded-gray-outline">
    {{template "poll-choice" (dict "label" .Choice1 "votes" .Choice1_Votes "poll" .)}}
    {{template "poll-choice" (dict "label" .Choice2 "votes" .Choice2_Votes "poll" .)}}
    {{if (gt .NumChoices 2)}}
      {{template "poll-choice" (dict "label" .Choice3 "votes" .Choice3_Votes "poll" .)}}
    {{end}}
    {{if (gt .NumChoices 3)}}
      {{template "poll-choice" (dict "label" .Choice4 "votes" .Choice4_Votes "poll" .)}}
    {{end}}

    <p class="poll__metadata">
      <span class="poll__metadata__state">
        {{if .IsOpen}}
          Poll open, voting ends at {{.FormatEndsAt}}
        {{else}}
          Poll ended {{.FormatEndsAt}}
        {{end}}
      </span>
      -
      <span class="poll-vote-count">{{.TotalVotes}} votes</span>
    </p>
  </div>
{{end}}
