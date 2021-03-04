# Windows Specific Documentation

Everything here should be considered pre-release and subject to change. Better support is incoming.

## Running as a Windows Service

### Config File 

Config File is passed via registry entry located at `Computer\HKEY_LOCAL_MACHINE\SYSTEM\ControlSet001\Services\<servicename>\ImagePath`
for example `C:\Program Files\Agent\agent.exe -config.file="C:\Program Files\Agent\agent-config.yaml"` This
will get easier once installer is finished.

### Logs

Currently, log messages are not recorded, there will be future work to write to the Windows Event Log.



