param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Files,

    [string]$Prefix = "Update",

    [switch]$WhatIf
)

$ErrorActionPreference = "Stop"

function Get-ChangedFiles {
    git ls-files --modified --others --exclude-standard
    if ($LASTEXITCODE -ne 0) {
        throw "git ls-files failed with exit code $LASTEXITCODE"
    }
}

function Get-CommitSubject([string]$Path) {
    $normalized = $Path.Replace("\", "/")
    return "$Prefix $normalized"
}

git rev-parse --is-inside-work-tree 2>$null | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw "This script must be run inside a git repository."
}

if (-not $Files -or $Files.Count -eq 0) {
    $Files = @(Get-ChangedFiles)
}

if (-not $Files -or $Files.Count -eq 0) {
    Write-Host "No changed files to commit."
    exit 0
}

foreach ($file in $Files) {
    if ([string]::IsNullOrWhiteSpace($file)) {
        continue
    }

    $subject = Get-CommitSubject $file

    if ($WhatIf) {
        Write-Host "Would commit: $file"
        Write-Host "  $subject"
        continue
    }

    git add -- "$file"
    if ($LASTEXITCODE -ne 0) {
        throw "git add failed for $file with exit code $LASTEXITCODE"
    }

    $staged = git diff --cached --name-only -- "$file"
    if ($LASTEXITCODE -ne 0) {
        throw "git diff --cached failed for $file with exit code $LASTEXITCODE"
    }
    if ([string]::IsNullOrWhiteSpace($staged)) {
        Write-Host "Skipping unchanged file: $file"
        continue
    }

    git commit -m "$subject" -- "$file"
    if ($LASTEXITCODE -ne 0) {
        throw "git commit failed for $file with exit code $LASTEXITCODE"
    }
}
