# Installs compress-env on Windows.
#
#   irm https://raw.githubusercontent.com/cavaldos/compress-env/main/scripts/install.ps1 | iex
#
# Environment:
#   VERSION      release to install, e.g. v0.0.1 (default: latest)
#   INSTALL_DIR  where to put the binary (default: %LOCALAPPDATA%\Programs\compress-env)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$repo = 'cavaldos/compress-env'
$base = if ($env:COMPRESS_ENV_BASE_URL) { $env:COMPRESS_ENV_BASE_URL } else { "https://github.com/$repo/releases" }
$version = if ($env:VERSION) { $env:VERSION } else { 'latest' }

switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { $arch = 'amd64' }
    'ARM64' { $arch = 'arm64' }
    default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

$dir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\compress-env' }
$file = "compress-env-windows-$arch.exe"
$url = if ($version -eq 'latest') { "$base/latest/download" } else { "$base/download/$version" }

$tmp = Join-Path ([IO.Path]::GetTempPath()) ([IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    Write-Host "Downloading $file ($version)..."
    Invoke-WebRequest -UseBasicParsing "$url/$file" -OutFile (Join-Path $tmp $file)
    Invoke-WebRequest -UseBasicParsing "$url/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt')

    $line = Get-Content (Join-Path $tmp 'checksums.txt') | Where-Object { $_ -match " $([regex]::Escape($file))$" } | Select-Object -First 1
    $expected = if ($line) { ($line -split '\s+')[0] } else { '' }
    $actual = (Get-FileHash -Algorithm SHA256 (Join-Path $tmp $file)).Hash.ToLower()
    if (-not $expected -or $expected -ne $actual) { throw 'Checksum mismatch, not installing' }

    New-Item -ItemType Directory -Force -Path $dir | Out-Null
    Move-Item -Force (Join-Path $tmp $file) (Join-Path $dir 'compress-env.exe')
}
finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

Write-Host "Installed: $dir\compress-env.exe"

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $dir) {
    [Environment]::SetEnvironmentVariable('Path', "$userPath;$dir", 'User')
    Write-Host "Added $dir to your PATH. Open a new terminal to use compress-env."
}
