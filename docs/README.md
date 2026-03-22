# Recipes Docs

This folder contains project decisions that should be treated as source-of-truth references for implementation and tests.

## Core Decision Docs

- `search-contract.md` - Exact behavior for strict/inclusive ingredient search, ranking, and response shape.
- `auth-security-baseline.md` - Authentication approach, authorization rules, and minimum security controls.
- `llm-fallback-contract.md` - DB-first search policy, fallback triggers, generation schema, and persistence rules.
- `architecture.md` - Repository structure, service map, development quick-start, and known issues.
- `domain-language.md` - Entity definitions, service boundaries, domain events, and language rules.

## Usage

1. Implement features to match these contracts.
2. Write tests directly against these contracts.
3. Update docs first when behavior changes, then update code/tests.

## Model Guidance For Decision Docs

- Primary: `GPT-5.3 Codex Spark`
- Fallback: `GPT-5.2 Codex`
- Backup: `GPT-5.1 Codex`
