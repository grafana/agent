#Requires -RunAsAdministrator

# Script to install Grafana agent for Windows
[CmdletBinding()]
param (
    [Parameter(Mandatory = $true, ValueFromPipelineByPropertyName)]
    [ValidateNotNullOrEmpty()]
    [string] $GCLOUD_HOSTED_METRICS_URL,

    [Parameter(Mandatory = $true, ValueFromPipelineByPropertyName)]
    [ValidateNotNullOrEmpty()]
    [string] $GCLOUD_HOSTED_METRICS_ID,

    [Parameter(Mandatory = $true, ValueFromPipelineByPropertyName)]
    [ValidateNotNullOrEmpty()]
    [string] $GCLOUD_SCRAPE_INTERVAL,

    [Parameter(Mandatory = $true, ValueFromPipelineByPropertyName)]
    [ValidateNotNullOrEmpty()]
    [string] $GCLOUD_HOSTED_LOGS_URL,

    [Parameter(Mandatory = $true, ValueFromPipelineByPropertyName)]
    [ValidateNotNullOrEmpty()]
    [string] $GCLOUD_HOSTED_LOGS_ID,

    [Parameter(Mandatory = $true, ValueFromPipelineByPropertyName)]
    [ValidateNotNullOrEmpty()]
    [string] $GCLOUD_RW_API_KEY
)

Write-Host "Setting up Grafana agent"
Write-Host "GCLOUD_HOSTED_METRICS_URL:" $GCLOUD_HOSTED_METRICS_URL
Write-Host "GCLOUD_HOSTED_METRICS_ID:" $GCLOUD_HOSTED_METRICS_ID
Write-Host "GCLOUD_SCRAPE_INTERVAL:" $GCLOUD_SCRAPE_INTERVAL
Write-Host "GCLOUD_HOSTED_LOGS_URL:" $GCLOUD_HOSTED_LOGS_URL
Write-Host "GCLOUD_HOSTED_LOGS_ID:" $GCLOUD_HOSTED_LOGS_ID
Write-Host "GCLOUD_RW_API_KEY:" $GCLOUD_RW_API_KEY

Write-Host "Downloading Grafana agent Windows Installer"
$DOWNLOAD_URL = "https://github.com/grafana/agent/releases/latest/download/grafana-agent-installer.exe.zip"
$OUTPUT_ZIP_FILE = ".\grafana-agent-installer.exe.zip"
$WORKING_DIR = Get-Location
Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $OUTPUT_ZIP_FILE
Expand-Archive -LiteralPath $OUTPUT_ZIP_FILE -DestinationPath $WORKING_DIR.path

# Install Grafana agent in silent mode
Write-Host "Installing Grafana agent for Windows"
.\grafana-agent-installer.exe /S

Write-Host "Retrieving Grafana agent config"
$CONFIG_URI = "https://storage.googleapis.com/cloud-onboarding/agent/config/config.yaml"
$CONFIG_FILE = ".\grafana-agent.yaml"
Invoke-WebRequest -Uri $CONFIG_URI -Outfile $CONFIG_FILE

Write-Host "Updating agent config file"
$content = Get-Content $CONFIG_FILE
$content = $content.Replace("{GCLOUD_HOSTED_METRICS_URL}", $GCLOUD_HOSTED_METRICS_URL)
$content = $content.Replace("{GCLOUD_HOSTED_METRICS_ID}", $GCLOUD_HOSTED_METRICS_ID)
$content = $content.Replace("{GCLOUD_SCRAPE_INTERVAL}", $GCLOUD_SCRAPE_INTERVAL)
$content = $content.Replace("{GCLOUD_HOSTED_LOGS_URL}", $GCLOUD_HOSTED_LOGS_URL)
$content = $content.Replace("{GCLOUD_HOSTED_LOGS_ID}", $GCLOUD_HOSTED_LOGS_ID)
$content = $content.Replace("{GCLOUD_RW_API_KEY}", $GCLOUD_RW_API_KEY)
$content | Set-Content $CONFIG_FILE

Move-Item $config_file "C:\Program Files\Grafana Agent\agent-config.yaml" -force

# Wait for service to initialize after first install
Write-Host "Wait for Grafana agent service to initialize"
Start-Sleep -Seconds 5

# Restart Grafana agent to load new configuration
Write-Host "Restarting Grafana agent service"
Stop-Service "Grafana Agent"
Start-Service "Grafana Agent"

# Wait for service to startup after restart
Write-Host "Wait for Grafana service to initialize after restart"
Start-Sleep -Seconds 10

# Show Grafana agent service status
Get-Service "Grafana Agent"
