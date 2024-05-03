{{define "text-with-entities"}}
  {{range (splitList "\n" .)}}
    <p class="text" hx-trigger="click consume">
      {{range (get_entities .)}}
        {{if (eq .EntityType 1)}}
          <!-- Mention -->
          <a class="entity" href="/{{.Contents}}">@{{.Contents}}</a>
        {{else if (eq .EntityType 2)}}
          <!-- Hashtag -->
          <a class="entity" href="/search/%23{{.Contents}}">#{{.Contents}}</a>
        {{else}}
          <!-- Just text -->
          <!-- TODO: Fix extra spaces being inserted between entities and text
               - e.g., `(@asdf)` renders as `( @asdf )`
               - https://css-tricks.com/fighting-the-space-between-inline-block-elements/
          -->

          {{.Contents}}
        {{end}}
      {{end}}
    </p>
  {{end}}
{{end}}
