'use client'

import { useMemo, useState } from 'react'

const QUICK_SUGGESTIONS = [
  'chicken',
  'rice',
  'garlic',
  'onion',
  'egg',
  'tomato',
  'spinach',
  'lemon',
]

function normalizeToken(value) {
  return String(value || '').trim().toLowerCase()
}

function parseIngredients(raw) {
  return String(raw || '')
    .split(',')
    .map((item) => normalizeToken(item))
    .filter(Boolean)
}

export default function PantryComposer({ defaultRaw = '' }) {
  const [tokens, setTokens] = useState(parseIngredients(defaultRaw))
  const [draft, setDraft] = useState('')

  const serialized = useMemo(() => tokens.join(', '), [tokens])

  function addToken(rawValue) {
    const value = normalizeToken(rawValue)
    if (!value) return
    setTokens((prev) => (prev.includes(value) ? prev : [...prev, value]))
    setDraft('')
  }

  function removeToken(value) {
    setTokens((prev) => prev.filter((item) => item !== value))
  }

  function onDraftKeyDown(event) {
    if (event.key === 'Enter' || event.key === ',') {
      event.preventDefault()
      addToken(draft)
    }
    if (event.key === 'Backspace' && !draft && tokens.length > 0) {
      event.preventDefault()
      setTokens((prev) => prev.slice(0, prev.length - 1))
    }
  }

  return (
    <div className="pantry-composer">
      <label htmlFor="pantryDraft" style={{ display: 'block', marginBottom: '0.35rem' }}>
        Ingredients
      </label>
      <p id="pantryComposerHint" className="muted" style={{ fontSize: '0.82rem' }}>
        Press Enter or comma to add items. Backspace removes the last item when the field is empty.
      </p>
      <input
        id="pantryDraft"
        type="text"
        placeholder="Type ingredient and press Enter"
        value={draft}
        onChange={(event) => setDraft(event.target.value)}
        onKeyDown={onDraftKeyDown}
        aria-describedby="pantryComposerHint"
      />
      <input type="hidden" name="ingredients" value={serialized} />

      {tokens.length > 0 && (
        <div className="chip-list" style={{ marginTop: '0.55rem' }}>
          {tokens.map((token) => (
            <span className="chip" key={token}>
              {token}
              <button
                type="button"
                aria-label={`Remove ${token}`}
                onClick={() => removeToken(token)}
              >
                ×
              </button>
            </span>
          ))}
        </div>
      )}

      <div style={{ display: 'flex', gap: '0.4rem', flexWrap: 'wrap', marginTop: '0.6rem' }} aria-label="Quick suggestions" role="group">
        {QUICK_SUGGESTIONS.map((item) => (
          <button
            type="button"
            key={item}
            className="btn btn-chip"
            onClick={() => addToken(item)}
            disabled={tokens.includes(item)}
          >
            + {item}
          </button>
        ))}
      </div>
    </div>
  )
}
