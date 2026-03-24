# Native Accessibility Manual Test Sheet v1

Use this sheet for each release candidate. Mark pass/fail and capture defects with reproduction steps.

## Run Metadata

- Release candidate:
- Tester:
- Date:
- Device(s):
- OS version(s):
- Screen reader mode: TalkBack / VoiceOver

## Global Checks

- [ ] Focus order follows visual order and is predictable.
- [ ] Every actionable control has a clear label.
- [ ] Dynamic text scaling does not clip or overlap key actions.
- [ ] Color contrast is readable for primary and secondary text.
- [ ] Error and success status updates are announced.

## Search Tab

- [ ] "Search recipes" heading is announced as a heading.
- [ ] Ingredients input label is announced clearly.
- [ ] Strict/inclusive mode controls are distinguishable by label and state.
- [ ] Complex toggle label and state are announced.
- [ ] Running search announces result count or error message.

## Favorites Tab

- [ ] User ID, token, and recipe ID fields are labeled.
- [ ] Token field is secure/masked in UI.
- [ ] Load/Sync/Add/Remove controls announce meaningful labels.
- [ ] Queue sync outcome announces replayed and pending counts.
- [ ] Error state announcement is clear and non-duplicative.

## Account Tab

- [ ] Username/email and password fields are clearly announced.
- [ ] Login/Refresh/List Sessions/Logout actions are distinguishable.
- [ ] Session cards are readable in sequence.
- [ ] Auth success/error states announce correctly.

## Defects

- Defect ID:
  - Screen:
  - Reproduction steps:
  - Expected:
  - Actual:
  - Severity:

## Signoff

- [ ] Android accessibility pass complete for this RC.
- [ ] iOS accessibility pass complete for this RC.
- [ ] Blocking accessibility defects resolved or waived with rationale.
