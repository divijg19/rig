$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$Repo = 'divijg19/rig'
$Version = if ($env:RIG_VERSION) { $env:RIG_VERSION } else { 'latest' }

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default { throw "unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)" }
}

$asset = "rig_windows_${arch}.zip"
if ($Version -eq 'latest') {
    $url = "https://github.com/$Repo/releases/latest/download/$asset"
} else {
    $url = "https://github.com/$Repo/releases/download/$Version/$asset"
}

$installDir = Join-Path $env:LOCALAPPDATA 'Programs\rig'
New-Item -ItemType Directory -Path $installDir -Force | Out-Null

$tempRoot = Join-Path $env:TEMP ("rig-install-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null

try {
    $zipPath = Join-Path $tempRoot $asset
    $extractDir = Join-Path $tempRoot 'extract'

    Invoke-WebRequest -Uri $url -OutFile $zipPath
    Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

    $rigExe = Get-ChildItem -Path $extractDir -Recurse -File | Where-Object { $_.Name -eq 'rig.exe' } | Select-Object -First 1
    if (-not $rigExe) {
        throw 'archive did not contain rig.exe'
    }

    $targetExe = Join-Path $installDir 'rig.exe'
    Copy-Item -Path $rigExe.FullName -Destination $targetExe -Force

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (-not $userPath) { $userPath = '' }

    $parts = @()
    if ($userPath -ne '') {
        $parts = $userPath -split ';' | Where-Object { $_ -ne '' }
    }

    $hasInstallDir = $false
    foreach ($p in $parts) {
        if ($p.TrimEnd('\\') -ieq $installDir.TrimEnd('\\')) {
            $hasInstallDir = $true
            break
        }
    }

    if (-not $hasInstallDir) {
        $newUserPath = if ($userPath -eq '') { $installDir } else { "$userPath;$installDir" }
        [Environment]::SetEnvironmentVariable('Path', $newUserPath, 'User')
    }

    Write-Output "Installed: $targetExe"
    Write-Output 'Restart your terminal to use rig.'
    Write-Output 'Use:'
    Write-Output '  rig run'
    Write-Output '  rig check'
    Write-Output '  rig dev'
}
finally {
    if (Test-Path $tempRoot) {
        Remove-Item -Path $tempRoot -Recurse -Force
    }
}
