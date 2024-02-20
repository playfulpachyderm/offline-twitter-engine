#define NAME "Offline Twitter"
#define EXE_NAME "twitter.exe"

[Setup]

AppName={#NAME}
AppVersion={#VERSION}
WizardStyle=modern
DefaultDirName={autopf}/offline-twitter
DefaultGroupName={#NAME}
UninstallDisplayIcon={app}/{#EXE_NAME}
PrivilegesRequiredOverridesAllowed=dialog



[Tasks]

Name: createdesktopshortcut; Description: "Create a &desktop shortcut"; GroupDescription: "Shortcuts"; Flags: unchecked
Name: createstartmenushortcut; Description: "Create a Start Menu entry"; GroupDescription: "Shortcuts"


[Files]

Source: "{#EXE_PATH}"; DestDir: "{app}"; Flags: recursesubdirs



[Icons]

Name: "{group}\{#NAME}"; Filename: "{app}\{#EXE_NAME}"; WorkingDir: "{commondocs}"; Tasks: createstartmenushortcut; Parameters: "webserver --auto-open --default-profile"
Name: "{group}\Uninstall {#NAME}"; Filename: "{uninstallexe}"; Tasks: createstartmenushortcut
Name: "{commondesktop}\{#NAME}"; Filename: "{app}\{#EXE_NAME}"; WorkingDir: "{commondocs}"; Tasks: createdesktopshortcut; Parameters: "webserver --auto-open --default-profile"
