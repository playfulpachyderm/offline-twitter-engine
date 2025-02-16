{{define "circle-profile-img"}}
  <a class="profile-image" href="/{{.Handle}}">
    {{/* TODO: add `width` and `height` attrs to the <img>*/}}
    <img class="profile-image__image"
      {{if .IsContentDownloaded}}
        src="/content/{{.GetProfileImageLocalPath}}"
      {{else}}
        src="{{.ProfileImageUrl}}"
      {{end}}
    >
  </a>
{{end}}

{{define "circle-profile-img-no-link"}}
  <a class="profile-image"
  	hx-trigger="click consume"
    onclick="image_carousel.querySelector('img').src = this.querySelector('img').src; image_carousel.showModal();"
  >
    {{/* TODO: add `width` and `height` attrs to the <img>*/}}
    <img class="profile-image__image"
      {{if .IsContentDownloaded}}
        src="/content/{{.GetProfileImageLocalPath}}"
      {{else}}
        src="{{.ProfileImageUrl}}"
      {{end}}
    >
  </a>
{{end}}
