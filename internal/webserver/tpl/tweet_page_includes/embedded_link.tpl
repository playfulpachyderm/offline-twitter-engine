{{define "embedded-link"}}
  <a
    class="embedded-link rounded-gray-outline"
    target="_blank"
    href="{{.Text}}"
    style="max-width: {{if (ne .ThumbnailWidth 0)}}{{.ThumbnailWidth}}px {{else}}fit-content {{end}}"
  >
    <img
      {{if .IsContentDownloaded}}
        src="/content/link_preview_images/{{.ThumbnailLocalPath}}"
      {{else}}
        src="{{.ThumbnailRemoteUrl}}"
      {{end}}
      class="embedded-link__preview-image"
      width="{{.ThumbnailWidth}}" height="{{.ThumbnailHeight}}"
    />
    <h3 class="embedded-link__title">{{.Title}}</h3>
    <p class="embedded-link__description">{{.Description}}</p>
    <span class="row embedded-link__domain">
      <img class="svg-icon" src="/static/icons/link3.svg" width="24" height="24" />
      <span class="embedded-link__domain__contents">{{(.GetDomain)}}</span>
    </span>
  </a>
{{end}}
