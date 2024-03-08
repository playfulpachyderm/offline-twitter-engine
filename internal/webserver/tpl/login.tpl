{{define "title"}}Login{{end}}

{{define "main"}}
<div class="login">
  <form hx-post="/change-session" hx-target=".nav-sidebar" hx-swap="outerHTML" hx-ext="json-enc">
    <label for="select-account">Choose account:</label>
    <select name="account" id="select-account">
      {{range .ExistingSessions}}
        <option value="{{.}}">@{{.}}</option>
      {{end}}
      <option value="no account">[no account (don't log in)]</option>
    </select>
    <div class="field-container submit-container">
      <input type='submit' value='Use account'>
    </div>
  </form>

  <p>Or log in</p>

  <form class="login-form" hx-post="/login" hx-target="body" hx-ext="json-enc">
    <div class="field-container">
      <label>Username</label>
      {{with .FormErrors.username}}
        <label class='error'>({{.}})</label>
      {{end}}
      <input name='username' value='{{.Username}}'>
    </div>
    <div class="field-container">
      <label>Password:</label>
      {{with .FormErrors.password}}
        <label class='error'>({{.}})</label>
      {{end}}
      <input type='password' name='password'>
    </div>
    <div class="field-container submit-container">
      <input type='submit' value='Login'>
    </div>

    <div class="htmx-spinner-container">
      <div class="htmx-spinner-background"></div>
      <img class="svg-icon htmx-spinner" src="/static/icons/spinner.svg" />
    </div>
  </form>
</div>
{{end}}
