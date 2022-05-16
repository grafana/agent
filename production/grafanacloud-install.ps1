# Script to install Grafana agen for Windows
param ($GCLOUD_STACK_ID, $GCLOUD_API_KEY, $GCLOUD_API_URL)

Write-Host "Setting up Grafana agent"

if ( -Not [bool](([System.Security.Principal.WindowsIdentity]::GetCurrent()).groups -match "S-1-5-32-544") ) {
	Write-Host "ERROR: The script needs to be run with Administrator privileges"
	exit
}

# Check if required parameters are present
if ($GCLOUD_STACK_ID -eq "") {
	Write-Host "ERROR: Required argument GCLOUD_STACK_ID missing"
	exit
}

if ($GCLOUD_API_KEY -eq "") {
	Write-Host "ERROR:  Required argument GCLOUD_API_KEY missing"
	exit
}

if ($GCLOUD_API_URL -eq "") {
	Write-Host "ERROR: Required argument GCLOUD_API_URL missing"
	exit
}

Write-Host "GCLOUD_STACK_ID:" $GCLOUD_STACK_ID
Write-Host "GCLOUD_API_KEY:" $GCLOUD_API_KEY
Write-Host "GCLOUD_API_URL:" $GCLOUD_API_URL

# Install Module Powershell-yaml required to convert agent config from json to yaml
Write-Host "Checking and installing required Powershell-yaml module"
Install-Module PowerShell-yaml

Write-Host "Downloading Grafana agent Windows Installer"
$DOWLOAD_URL = "https://github.com/grafana/agent/releases/latest/download/grafana-agent-installer.exe"
$OUTPUT_FILE = ".\grafana-agent-installer.exe"
Invoke-WebRequest -Uri $DOWLOAD_URL -OutFile $OUTPUT_FILE

# Install Grafana agent in silent mode
Write-Host "Installing Grafana agent for Windows"
.\grafana-agent-installer.exe /S

Write-Host "Retrieving and updating Grafana agent config"
$CONFIG_URI = "$GCLOUD_API_URL/stacks/$GCLOUD_STACK_ID/agent_config"
$AUTH_TOKEN = "Bearer $GCLOUD_API_KEY"

$headers = New-Object "System.Collections.Generic.Dictionary[[String],[String]]"
$headers.Add("Authorization", $AUTH_TOKEN)

$response = Invoke-WebRequest $CONFIG_URI -Method 'GET' -Headers $headers

$jsonObj = $response | ConvertFrom-Json
if ($jsonObj.status -eq "success") {
	Write-Host "Saving and updatig agent configuration file"
	$yamlConfig = $jsonObj.data | ConvertTo-Yaml
	Set-Content -Path ".\agent-config.yaml" -Value ($yamlConfig)
	Move-Item ".\agent-config.yaml" "C:\Program Files\Grafana Agent\agent-config.yaml" -force
	Write-Host "Wait for Grafana service to initialize"

	# Wait for service to initialize after first install
	Start-Sleep -s 5

	# Restart Grafana agent to load new configuration
	Write-Host "Restarting Grafana agent service"
	Stop-Service "Grafana Agent"
	Start-Service "Grafana Agent"

	# Show Grafana agent service status
	Get-Service "Grafana Agent"
} else {
	Write-Host "Failed to retrieve config"
}