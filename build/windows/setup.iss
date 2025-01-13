#define NAME "Offline Twitter"
#define EXE_NAME "twitter.exe"
; The `version` macro should be passed from command line using `/Dversion=[...]`

[Setup]

AppName={#NAME}
AppVersion={#version}
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

Name: "{group}\{#NAME}"; Filename: "{app}\{#EXE_NAME}"; WorkingDir: "{autodocs}"; Tasks: createstartmenushortcut; Parameters: "--default-profile webserver --auto-open"
Name: "{group}\Uninstall {#NAME}"; Filename: "{uninstallexe}"; Tasks: createstartmenushortcut
Name: "{autodesktop}\{#NAME}"; Filename: "{app}\{#EXE_NAME}"; WorkingDir: "{autodocs}"; Tasks: createdesktopshortcut; Parameters: "--default-profile webserver --auto-open"

; [Registry]
; Root: HKCU; Subkey: "Environment"; ValueType: string; ValueName: "Path"; ValueData: "{olddata};{app}";
