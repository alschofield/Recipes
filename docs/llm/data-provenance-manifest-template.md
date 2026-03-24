# LLM Data Provenance Manifest Template

Record every dataset or corpus segment used for training/fine-tuning/eval.

## Source Record

- Source name:
- Source URL:
- Retrieved by:
- Retrieved date (UTC):
- Content type (recipes, ingredients, metadata):

## Rights and Licensing

- License type:
- Commercial use allowed: [ ]
- Derivative works allowed: [ ]
- Attribution required: [ ]
- Share-alike required: [ ]
- Additional restrictions:
- Legal review ticket/reference:

## Processing Pipeline

- Raw snapshot location:
- Cleaning script/version:
- Dedup method:
- Safety filtering method:
- Final artifact location:

## Compliance Decision

- Approved for eval only: [ ]
- Approved for fine-tuning: [ ]
- Approved for production inference context: [ ]
- Approved by:
- Approval date:
- Expiration/re-review date:

## Repo implementation references

- Example manifest: `llm/train/datasets/provenance-manifest.v1.json`
- Validation script: `llm/train/datasets/validate_provenance.py`
- Denylist: `llm/train/datasets/source-denylist.txt`
