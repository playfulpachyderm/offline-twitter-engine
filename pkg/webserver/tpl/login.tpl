{{define "title"}}Login{{end}}

{{define "main"}}
<div class="login-page">
  <h1>Login</h1>

  <form class="choose-session" hx-post="/change-session" hx-target="#nav-sidebar" hx-swap="outerHTML" hx-ext="json-enc">
    <h3>Open existing session</h3>
    <div class="row row--spread choose-session__form-contents">
      <select name="account" class="choose-session__dropdown">
        {{range .ExistingSessions}}
          <option value="{{.}}">@{{.}}</option>
        {{end}}
        <option value="no account">[no account (don't log in)]</option>
      </select>
      <div class="login-form__field-container login-form__submit-container">
        <input type='submit' value='Go'>
      </div>
    </div>
  </form>

  <hr>

  <form class="login-form" hx-post="/login" hx-target="body" hx-ext="json-enc">
    <h3>Log in (new session)</h3>
    <div class="login-form__field-container">
      <label>Username</label>
      {{with .FormErrors.username}}
        <label class='login-form__error-label'>({{.}})</label>
      {{end}}
      <input name='username' value='{{.Username}}' class="login-form__input">
    </div>
    <div class="login-form__field-container">
      <label>Password:</label>
      {{with .FormErrors.password}}
        <label class='login-form__error-label'>({{.}})</label>
      {{end}}
      <input type='password' name='password' class="login-form__input">
    </div>
    <div class="login-form__field-container login-form__submit-container">
      <input type='submit' value='Login'>
    </div>

    <div class="htmx-spinner">
      <div class="htmx-spinner__background"></div>
      <img class="svg-icon htmx-spinner__icon" src="/static/icons/spinner.svg" />
    </div>
  </form>
</div>
{{end}}
