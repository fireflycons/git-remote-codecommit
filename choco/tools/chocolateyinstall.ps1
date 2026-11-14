$ErrorActionPreference = 'Stop'

$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Definition

$isArm64 = $env:PROCESSOR_ARCHITECTURE -eq 'ARM64' -or $env:PROCESSOR_ARCHITEW6432 -eq 'ARM64'
$unused = if ($isArm64) { 'amd64' } else { 'arm64' }

# Both architectures are embedded; remove the unused one so Chocolatey
# creates a single git-remote-codecommit shim.
Remove-Item -Recurse -Force (Join-Path $toolsDir $unused)
