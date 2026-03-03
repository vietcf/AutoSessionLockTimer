[Setup]
AppName=AutoLock Session Timer
AppVersion=1.0.0
AppPublisher=AutoLock
DefaultDirName={localappdata}\AutoLock
DefaultGroupName=AutoLock
UninstallDisplayIcon={app}\AutoLock.exe
OutputDir=.
OutputBaseFilename=AutoLockSetup
Compression=lzma
SolidCompression=yes
ArchitecturesInstallIn64BitMode=x64
PrivilegesRequired=lowest

[Files]
Source: "..\AutoLock.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\config.json"; DestDir: "{app}"; Flags: onlyifdoesntexist

[Icons]
Name: "{group}\AutoLock Session Timer"; Filename: "{app}\AutoLock.exe"
Name: "{group}\Uninstall AutoLock"; Filename: "{uninstallexe}"

[Run]
Filename: "{app}\AutoLock.exe"; Description: "Run AutoLock now"; Flags: nowait postinstall skipifsilent
