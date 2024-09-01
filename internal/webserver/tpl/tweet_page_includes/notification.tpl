{{define "notification"}}
  {{$notification := (notification .NotificationID)}}

  <div class="notification">
    <div class="notification__header">
      {{if (not (eq $notification.ActionUserID 0))}}
        <div class="notification__users">
          {{template "circle-profile-img" (user $notification.ActionUserID)}}
          {{/*template "author-info" (user $notification.ActionUserID)*/}}
          {{if (gt (len $notification.UserIDs) 1)}}
            {{$max_display_users := 10}}
            {{range $i, $user_id := $notification.UserIDs}}
              {{if (ne $user_id $notification.ActionUserID)}} {{/* don't duplicate main user */}}
                {{/* Only render the first 10-ish users */}}
                {{if (lt $i $max_display_users)}}
                  {{template "circle-profile-img" (user $user_id)}}
                {{end}}
              {{end}}
            {{end}}
            {{if (gt (len $notification.UserIDs) (add $max_display_users 1))}}
              <span class="ellipsis">...</span>
            {{end}}
          {{end}}
        </div>
      {{end}}

      <div class="notification__text">
        {{if (eq $notification.Type 1)}} {{/* LIKE */}}
          <b>{{(user $notification.ActionUserID).DisplayName}} liked your tweet</b>
        {{else if (eq $notification.Type 2)}} {{/* RETWEET */}}
          <b>{{(user $notification.ActionUserID).DisplayName}} retweeted you</b>
        {{else if (eq $notification.Type 3)}} {{/* QUOTE_TWEET */}}
          <b>{{(user $notification.ActionUserID).DisplayName}} quote-tweeted you</b>
        {{else if (eq $notification.Type 4)}} {{/* REPLY */}}
          <b>{{(user $notification.ActionUserID).DisplayName}} replied to you</b>
        {{else if (eq $notification.Type 5)}} {{/* FOLLOW */}}
          <b>{{(user $notification.ActionUserID).DisplayName}} followed you!</b>
        {{else if (eq $notification.Type 6)}} {{/* MENTION */}}
          <b>{{(user $notification.ActionUserID).DisplayName}} mentioned you</b>
        {{else if (eq $notification.Type 7)}} {{/* USER_IS_LIVE */}}
          <b>{{(user $notification.ActionUserID).DisplayName}} is live</b>
        {{else if (eq $notification.Type 8)}} {{/* POLL_ENDED */}}
          <b>Poll ended.</b>
        {{else if (eq $notification.Type 9)}} {{/* LOGIN */}}
          <b>New login on your account.</b>
        {{else if (eq $notification.Type 10)}} {{/* COMMUNITY_PINNED_POST */}}
          <b>{{(user $notification.ActionUserID).DisplayName}} posted in community</b>
        {{else if (eq $notification.Type 11)}} {{/* RECOMMENDED_POST */}}
          <b>You've been recommended a post from {{(user $notification.ActionUserID).DisplayName}}</b>
        {{else}}
          <b>{{"<<UNKNOWN ID>>: "}}{{$notification.Type}}</b>
        {{end}}
      </div>
    </div>

    {{if (ne .TweetID 0)}}
      {{template "tweet" .}}
    {{end}}
  </div>
{{end}}
