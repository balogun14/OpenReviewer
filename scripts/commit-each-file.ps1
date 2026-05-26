param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$Files,

    [string]$Prefix = "Update",

    [switch]$WhatIf
)

$ErrorActionPreference = "Stop"

function Get-ChangedFiles {
    git ls-files --modified --others --exclude-standard
}

function Get-CommitSubject([string]$Path) {
    $normalized = $Path.Replace("\", "/")
    return "$Prefix $normalized"
}

if (-not (git rev-parse --is-inside-work-tree 2>$null)) {
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

    $staged = git diff --cached --name-only -- "$file"
    if ([string]::IsNullOrWhiteSpace($staged)) {
        Write-Host "Skipping unchanged file: $file"
        continue
    }

    git commit -m "$subject" -- "$file"
}
