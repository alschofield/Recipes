# Dataset Review and Consolidation Notes

Generated from files in `datasets/raw/server-lib`.

## What each file is

- `flavourDB2.json` (120 MB, 935 records)
  - Flavor chemistry knowledge base by ingredient/entity.
  - Key fields include ingredient identity/category plus `molecules` arrays with aroma/flavor metadata.
  - Strong candidate for ingredient quality signals and flavor pairing logic.

- `food_info.csv` (51 KB, 935 rows)
  - Compact index that aligns 1:1 with `flavourDB2.json` by ID and food name.
  - Contains light metadata (`food_name`, `category`, `natural_source`, `synonyms`, `molecule_count`).
  - Best used as a quick lookup table; likely derived from flavourDB.

- `CulinaryDB.zip` (5.1 MB)
  - `01_Recipe_Details.csv` (45,772 recipes, title/source/cuisine)
  - `02_Ingredients.csv` (930 canonical ingredient aliases/synonyms/categories)
  - `03_Compound_Ingredients.csv` (103 blends like garam masala)
  - `04_Recipe-Ingredients_Aliases.csv` (456,279 recipe-to-ingredient alias mappings)
  - Best source for alias normalization and ingredient canonicalization.

- `kaggle ingredients dataset.zip` (2.3 MB)
  - `train.json` (39,774 recipes with `cuisine` + ingredient list)
  - `test.json` (9,944 recipes)
  - Strong signal for ingredient co-occurrence and cuisine prediction.

- `collection of recipes dataset from kaggle.zip` (9.6 KB)
  - Small synthetic/curated set (161 rows), includes prep/cook time, calories, diet tags.
  - Useful as schema example and UI/demo seed data.

- `FoodData_Central_csv_2025-12-18.zip` (458 MB compressed, ~3.25 GB uncompressed)
  - USDA FDC relational dump.
  - Core tables: `food.csv` (~2.0M), `food_nutrient.csv` (~26.2M), `nutrient.csv`, `branded_food.csv` (~1.95M), etc.
  - Authoritative nutrition data for quality scoring and nutrition constraints.

- `FoodData_Central_branded_food_json_2025-12-18.zip` (195 MB compressed, ~3.3 GB JSON)
  - USDA branded foods JSON variant.
  - Huge; overlaps with CSV source but less convenient for SQL-style joins.

- `FoodData_Central_foundation_food_json_2025-12-18.zip` (~0.5 MB compressed)
  - USDA Foundation foods JSON.

- `FoodData_Central_sr_legacy_food_json_2018-04.zip` (~13 MB compressed)
  - USDA SR Legacy foods JSON.

- `FoodData_Central_survey_food_json_2024-10-31.zip` (~3.7 MB compressed)
  - USDA Survey/FNDDS foods JSON.

- `indonesian spices dataset.zip` (402 MB, 6,510 images)
  - 31 classes x 210 images each (plus `bukan rempah` = non-spice class).
  - Computer-vision training data for spice recognition (not direct recipe seeding text).

## Similarities and overlap

- `food_info.csv` and `flavourDB2.json` are directly aligned (same 935 IDs/foods).
- Culinary/Kaggle/FlavourDB ingredient vocab overlaps are meaningful for canonical ingredient IDs.
- USDA data overlaps by ingredient name but is much broader/noisier (consumer packaged foods + variants).
- Indonesian spice images complement text datasets if you want photo-to-ingredient workflows.

## Consolidation output created now

- `unified_ingredient_index.csv`
- `unified_ingredient_index.summary.json`
- `unified_ingredient_index.zip`

Location: `datasets/derived/server-lib`

`unified_ingredient_index.csv` combines normalized ingredient names with source presence flags and signals:

- source coverage flags (`in_flavourdb`, `in_food_info`, `in_culinarydb`, `in_kaggle_train`, `in_world_recipes`)
- metadata (`flavourdb_category`, `food_info_category`, `food_info_natural_source`, `flavour_molecule_count`)
- usage frequencies (`culinary_alias_count`, `kaggle_recipe_count`, `world_recipe_count`)
- nutrition linkage signal (`usda_exact_name_matches` from `food.csv` exact normalized name matches)

Current size:

- 154,514 normalized ingredient rows
- 8.9 MB CSV
- 1.3 MB zipped artifact

## Practical project uses

- Ingredient normalization service
  - Map messy user input (`capsicum`, `green bell pepper`, etc.) to canonical ingredient IDs.
- Better seed quality
  - Prefer ingredients appearing across multiple datasets and with flavor/nutrition backing.
- Recipe quality scoring
  - Penalize low-nutrient/high-processed combinations using USDA nutrient joins.
  - Reward diversity, ingredient integrity, and flavor compatibility.
- Search and recommendation
  - Use cuisine-conditioned ingredient priors from Kaggle + CulinaryDB.
  - Use FlavourDB molecules to support pairing suggestions and substitution hints.
- Future CV features
  - Train spice detector from Indonesian dataset to pre-fill ingredient inputs from photos.

## Recommended next consolidation step

Build a versioned canonical seed artifact (for DB migrations/seed command) with:

1. `ingredients_canonical` (ID, canonical name, aliases, category)
2. `ingredient_nutrition_baseline` (selected macro/micro nutrients from USDA)
3. `ingredient_flavor_profile` (molecule counts + key descriptors from FlavourDB)
4. `ingredient_quality_score` (transparent formula and components)

Then publish as `canonical_ingredient_seed_v1.csv` + `.zip` under `datasets/derived/server-lib`.
