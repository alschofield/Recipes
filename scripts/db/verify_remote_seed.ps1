$ErrorActionPreference = 'Stop'

function Require-EnvVar {
  param([string]$Name)

  if ([string]::IsNullOrWhiteSpace((Get-Item -Path "Env:$Name" -ErrorAction SilentlyContinue).Value)) {
    throw "Environment variable '$Name' is required."
  }
}

function Require-Command {
  param([string]$Name)

  if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
    throw "Required command '$Name' was not found in PATH."
  }
}

function Invoke-ScalarQuery {
  param([string]$Sql)

  $result = (& psql $env:DATABASE_URL -v ON_ERROR_STOP=1 -t -A -c $Sql)
  if ($LASTEXITCODE -ne 0) {
    throw "Query failed: $Sql"
  }

  return ($result | Out-String).Trim()
}

function To-Int {
  param([string]$Value)

  $parsed = 0
  if (-not [int]::TryParse($Value, [ref]$parsed)) {
    throw "Expected integer query result but received '$Value'"
  }

  return $parsed
}

Require-EnvVar -Name 'DATABASE_URL'
Require-Command -Name 'psql'

$minIngredients = if ($env:SEED_MIN_INGREDIENTS) { [int]$env:SEED_MIN_INGREDIENTS } else { 50 }
$minRecipes = if ($env:SEED_MIN_RECIPES) { [int]$env:SEED_MIN_RECIPES } else { 3 }
$minAliases = if ($env:SEED_MIN_ALIASES) { [int]$env:SEED_MIN_ALIASES } else { 5 }

Write-Host '[db/verify] Running remote seed verification checks...'

$ingredientsCount = To-Int (Invoke-ScalarQuery "SELECT COUNT(*) FROM ingredients;")
$recipesCount = To-Int (Invoke-ScalarQuery "SELECT COUNT(*) FROM recipes;")
$aliasesCount = To-Int (Invoke-ScalarQuery "SELECT COUNT(*) FROM ingredient_aliases;")
$recipeIngredientsCount = To-Int (Invoke-ScalarQuery "SELECT COUNT(*) FROM recipe_ingredients;")

$recipesMissingName = To-Int (Invoke-ScalarQuery "SELECT COUNT(*) FROM recipes WHERE name IS NULL OR btrim(name) = ''; ")
$recipesMissingSteps = To-Int (Invoke-ScalarQuery "SELECT COUNT(*) FROM recipes WHERE steps IS NULL OR jsonb_typeof(steps) <> 'array' OR jsonb_array_length(steps) = 0;")
$ingredientsMissingCanonical = To-Int (Invoke-ScalarQuery "SELECT COUNT(*) FROM ingredients WHERE canonical_name IS NULL OR btrim(canonical_name) = ''; ")
$aliasesMissingValue = To-Int (Invoke-ScalarQuery "SELECT COUNT(*) FROM ingredient_aliases WHERE alias IS NULL OR btrim(alias) = ''; ")

$duplicateIngredientNames = To-Int (Invoke-ScalarQuery @"
SELECT COUNT(*)
FROM (
  SELECT lower(btrim(canonical_name))
  FROM ingredients
  GROUP BY lower(btrim(canonical_name))
  HAVING COUNT(*) > 1
) dup;
"@)

$duplicateAliases = To-Int (Invoke-ScalarQuery @"
SELECT COUNT(*)
FROM (
  SELECT lower(btrim(alias))
  FROM ingredient_aliases
  GROUP BY lower(btrim(alias))
  HAVING COUNT(*) > 1
) dup;
"@)

$recipesWithoutIngredients = To-Int (Invoke-ScalarQuery @"
SELECT COUNT(*)
FROM recipes r
LEFT JOIN recipe_ingredients ri ON ri.recipe_id = r.id
WHERE ri.id IS NULL;
"@)

$databaseSourceCount = To-Int (Invoke-ScalarQuery "SELECT COUNT(*) FROM recipes WHERE source = 'database';")
$llmSourceCount = To-Int (Invoke-ScalarQuery "SELECT COUNT(*) FROM recipes WHERE source = 'llm';")

$checks = @(
  @{ Name = 'ingredients_count_min'; Passed = ($ingredientsCount -ge $minIngredients); Detail = "actual=$ingredientsCount min=$minIngredients" },
  @{ Name = 'recipes_count_min'; Passed = ($recipesCount -ge $minRecipes); Detail = "actual=$recipesCount min=$minRecipes" },
  @{ Name = 'aliases_count_min'; Passed = ($aliasesCount -ge $minAliases); Detail = "actual=$aliasesCount min=$minAliases" },
  @{ Name = 'recipe_ingredients_nonzero'; Passed = ($recipeIngredientsCount -gt 0); Detail = "actual=$recipeIngredientsCount" },
  @{ Name = 'recipes_missing_name'; Passed = ($recipesMissingName -eq 0); Detail = "actual=$recipesMissingName" },
  @{ Name = 'recipes_missing_steps'; Passed = ($recipesMissingSteps -eq 0); Detail = "actual=$recipesMissingSteps" },
  @{ Name = 'ingredients_missing_canonical'; Passed = ($ingredientsMissingCanonical -eq 0); Detail = "actual=$ingredientsMissingCanonical" },
  @{ Name = 'aliases_missing_value'; Passed = ($aliasesMissingValue -eq 0); Detail = "actual=$aliasesMissingValue" },
  @{ Name = 'duplicate_ingredient_names'; Passed = ($duplicateIngredientNames -eq 0); Detail = "actual=$duplicateIngredientNames" },
  @{ Name = 'duplicate_aliases'; Passed = ($duplicateAliases -eq 0); Detail = "actual=$duplicateAliases" },
  @{ Name = 'recipes_without_ingredients'; Passed = ($recipesWithoutIngredients -eq 0); Detail = "actual=$recipesWithoutIngredients" },
  @{ Name = 'database_source_nonzero'; Passed = ($databaseSourceCount -gt 0); Detail = "actual=$databaseSourceCount" }
)

Write-Host ''
Write-Host '[db/verify] Metrics'
Write-Host "  ingredients=$ingredientsCount"
Write-Host "  recipes=$recipesCount"
Write-Host "  ingredient_aliases=$aliasesCount"
Write-Host "  recipe_ingredients=$recipeIngredientsCount"
Write-Host "  source.database=$databaseSourceCount"
Write-Host "  source.llm=$llmSourceCount"

Write-Host ''
Write-Host '[db/verify] Checks'

$failedChecks = @()
foreach ($check in $checks) {
  if ($check.Passed) {
    Write-Host "  PASS $($check.Name) ($($check.Detail))"
  } else {
    Write-Host "  FAIL $($check.Name) ($($check.Detail))"
    $failedChecks += $check
  }
}

Write-Host ''
if ($failedChecks.Count -gt 0) {
  Write-Host "[db/verify] Verification failed: $($failedChecks.Count) check(s) did not pass."
  exit 1
}

Write-Host '[db/verify] Verification passed.'
