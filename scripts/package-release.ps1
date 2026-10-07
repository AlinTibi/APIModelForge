param(
  [Parameter(Mandatory=$true)][ValidatePattern('^v\d+\.\d+\.\d+$')][string]$Tag,
  [string]$OutputDirectory = 'build/rc'
)
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path $PSScriptRoot -Parent
$metadata = Get-Content -Raw (Join-Path $repoRoot 'wails.json') | ConvertFrom-Json
if ($metadata.info.productVersion -ne $Tag.Substring(1)) { throw 'Tag does not match product version' }
$exePath = Join-Path $repoRoot 'build/bin/APIModelForge.exe'
$exeInfo = [Diagnostics.FileVersionInfo]::GetVersionInfo($exePath)
if ($exeInfo.ProductVersionRaw.ToString(3) -ne $Tag.Substring(1) -or $exeInfo.FileVersionRaw.ToString(3) -ne $Tag.Substring(1)) { throw 'Executable version does not match tag' }
$outputPath = [IO.Path]::GetFullPath((Join-Path $repoRoot $OutputDirectory))
New-Item -ItemType Directory -Force -Path $outputPath | Out-Null
$stagePath = Join-Path $outputPath ('stage-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $stagePath | Out-Null
try {
  Copy-Item -LiteralPath $exePath -Destination (Join-Path $stagePath 'APIModelForge.exe')
  foreach ($file in @('README.md','LICENSE')) { Copy-Item -LiteralPath (Join-Path $repoRoot $file) -Destination (Join-Path $stagePath $file) }
  $zipPath = Join-Path $outputPath "APIModelForge-$Tag-win-x64.zip"
  Compress-Archive -Path (Join-Path $stagePath '*') -DestinationPath $zipPath -CompressionLevel Optimal -Force
  $hash = (Get-FileHash -LiteralPath $zipPath -Algorithm SHA256).Hash.ToLowerInvariant()
  $checksumPath = "$zipPath.sha256"
  [IO.File]::WriteAllText($checksumPath,"$hash  $([IO.Path]::GetFileName($zipPath))`n",[Text.UTF8Encoding]::new($false))
  [pscustomobject]@{Zip=$zipPath;Checksum=$checksumPath;SHA256=$hash}
} finally {
  if ([IO.Path]::GetFullPath($stagePath).StartsWith($outputPath + [IO.Path]::DirectorySeparatorChar)) { Remove-Item -LiteralPath $stagePath -Recurse -Force }
}
