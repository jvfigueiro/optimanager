<#
.SYNOPSIS
  Configura o OptiManager no Windows: libera a porta no Firewall e inicia o servidor junto com o sistema.

.DESCRIPTION
  - Cria a regra de Firewall de entrada (TCP, porta 8585 por padrao) para acesso pela rede local.
  - Cria uma Tarefa Agendada "OptiManager" que inicia o executavel na inicializacao do Windows
    (conta SYSTEM, sem precisar de login), com reinicio automatico em caso de falha.
  - Idempotente: pode ser executado novamente sem duplicar regras.

.EXAMPLE
  # PowerShell como Administrador, na pasta onde esta o optimanager.exe:
  .\setup-optimanager.ps1

.EXAMPLE
  .\setup-optimanager.ps1 -Port 9000 -ExePath "C:\OptiManager\optimanager.exe"

.EXAMPLE
  .\setup-optimanager.ps1 -Uninstall
#>
[CmdletBinding()]
param(
  [string]$ExePath = "",
  [int]$Port = 8585,
  [switch]$Uninstall
)

$ErrorActionPreference = "Stop"
$TaskName = "OptiManager"
$RuleName = "OptiManager (Porta $Port)"

function Test-IsAdmin {
  $id = [Security.Principal.WindowsIdentity]::GetCurrent()
  return ([Security.Principal.WindowsPrincipal]$id).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

if (-not (Test-IsAdmin)) {
  Write-Host "Execute este script em um PowerShell como Administrador." -ForegroundColor Red
  Write-Host "Dica: se der erro de politica de execucao, use: powershell -ExecutionPolicy Bypass -File .\setup-optimanager.ps1"
  exit 1
}

if ($Uninstall) {
  Get-NetFirewallRule -DisplayName "OptiManager (Porta *)" -ErrorAction SilentlyContinue | Remove-NetFirewallRule
  if (Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue) {
    Stop-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
    Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
  }
  Write-Host "OptiManager removido da inicializacao e do Firewall. Os dados (pasta data) foram preservados." -ForegroundColor Green
  exit 0
}

# Localiza o executavel
if ([string]::IsNullOrWhiteSpace($ExePath)) {
  $here = Split-Path -Parent $MyInvocation.MyCommand.Path
  foreach ($candidate in @("optimanager.exe", "optimanager-windows-amd64.exe")) {
    $p = Join-Path $here $candidate
    if (Test-Path $p) { $ExePath = $p; break }
  }
}
if ([string]::IsNullOrWhiteSpace($ExePath) -or -not (Test-Path $ExePath)) {
  Write-Host "Executavel nao encontrado. Coloque este script na mesma pasta do optimanager.exe ou use -ExePath." -ForegroundColor Red
  exit 1
}
$ExePath = (Resolve-Path $ExePath).Path
$WorkDir = Split-Path -Parent $ExePath

# 1. Firewall
Get-NetFirewallRule -DisplayName $RuleName -ErrorAction SilentlyContinue | Remove-NetFirewallRule
New-NetFirewallRule -DisplayName $RuleName -Direction Inbound -Protocol TCP -LocalPort $Port -Action Allow -Profile Private,Domain | Out-Null
Write-Host "[OK] Firewall: porta $Port liberada (perfis Privado e Dominio)." -ForegroundColor Green

# 2. Inicio automatico com o Windows
if (Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue) {
  Stop-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
  Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
}

$action = New-ScheduledTaskAction -Execute $ExePath -WorkingDirectory $WorkDir
$trigger = New-ScheduledTaskTrigger -AtStartup
$principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount -RunLevel Highest
$settings = New-ScheduledTaskSettingsSet `
  -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries `
  -StartWhenAvailable `
  -RestartCount 5 -RestartInterval (New-TimeSpan -Minutes 1) `
  -ExecutionTimeLimit ([TimeSpan]::Zero) `
  -MultipleInstances IgnoreNew

Register-ScheduledTask -TaskName $TaskName -Action $action -Trigger $trigger -Principal $principal -Settings $settings `
  -Description "Servidor OptiManager (inicia com o Windows)" | Out-Null
Write-Host "[OK] Tarefa agendada '$TaskName' criada (inicia com o Windows)." -ForegroundColor Green

Start-ScheduledTask -TaskName $TaskName
Start-Sleep -Seconds 3

# 3. Resumo
$ips = Get-NetIPAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue |
  Where-Object { $_.IPAddress -notlike "127.*" -and $_.IPAddress -notlike "169.254.*" } |
  Select-Object -ExpandProperty IPAddress
Write-Host ""
Write-Host "OptiManager em execucao. Acesse:" -ForegroundColor Cyan
Write-Host "  Neste PC:       http://localhost:$Port"
foreach ($ip in $ips) { Write-Host "  Na rede local:  http://${ip}:$Port" }
Write-Host ""
Write-Host "Dados em: $(Join-Path $WorkDir 'data')  (faca backup periodico dessa pasta)"
Write-Host "Para remover: .\setup-optimanager.ps1 -Uninstall"
