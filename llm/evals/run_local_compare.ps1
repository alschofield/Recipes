param(
  [string]$BaseUrl = "http://localhost:11434/v1",
  [string]$ApiKey = "local-not-used",
  [string]$ModelA = "qwen3:8b",
  [string]$ModelB = "mistral:latest"
)

$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$recipeCases = Join-Path $root "llm\evals\recipe_cases.jsonl"
$safetyCases = Join-Path $root "llm\evals\safety_cases.jsonl"
$complexCases = Join-Path $root "llm\evals\complex_cases.jsonl"

function Run-Eval {
  param([string]$Model)

  $safeModel = $Model -replace "[/:]", "-"
  $outDir = Join-Path $root "llm\evals\results\$safeModel"

  python (Join-Path $root "llm\evals\run_eval.py") `
    --base-url $BaseUrl `
    --api-key $ApiKey `
    --model $Model `
    --recipe-cases $recipeCases `
    --safety-cases $safetyCases `
    --complex-cases $complexCases `
    --out $outDir
}

Run-Eval -Model $ModelA
Run-Eval -Model $ModelB

Write-Host "Finished local compare for: $ModelA and $ModelB"
Write-Host "Results folder: $(Join-Path $root 'llm\evals\results')"
