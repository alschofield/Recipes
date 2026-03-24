# V1 External Inputs Checklist

Track non-code values/credentials/approvals needed to close launch blockers.

## Identity and Domains

- [x] Final production web domain (`www`): **www.ingrediential.uk**
- [ ] Final production API domain (`api`): **api.ingrediential.uk (pending Worker route + origin wiring)**
- [x] Domain ownership access confirmed: **ingrediential.uk managed in Cloudflare**
- [x] API routing strategy selected: **Cloudflare Worker gateway**

## API Gateway Origins

- [x] `RECIPES_ORIGIN` value: `https://recipes-production-b30c.up.railway.app`
- [x] `USERS_ORIGIN` value: `https://users-production-8fab.up.railway.app`
- [x] `FAVORITES_ORIGIN` value: `https://favorites-production.up.railway.app`

## Contacts and Ownership

- [x] Technical release owner: **Alexander Schofield (alex.schofield816@gmail.com)**
- [x] Security/ops approver: **Alexander Schofield (alex.schofield816@gmail.com)**
- [x] Product owner signoff contact: **Alexander Schofield (alex.schofield816@gmail.com)**
- [ ] On-call/escalation channel: **TBD**
- [ ] Third-party escalation provider (if needed): **TBD**

## Secrets and Credentials

- [ ] Production JWT secret provisioned in secret manager.
- [ ] Production DB credentials provisioned.
- [ ] Production Redis credentials provisioned.
- [ ] Deploy hook secrets configured in GitHub.
- [ ] Mobile signing materials available (Android keystore, iOS signing cert/profile).

## Store and Policy Inputs

- [ ] Support email and support URL.
- [ ] Marketing URL.
- [ ] Privacy policy URL for web/mobile declarations.
- [ ] Play Console access and role assignments.
- [ ] App Store Connect/TestFlight access and role assignments.

## Approval Inputs

- [ ] Security/operations final approval date:
- [ ] Product launch approval date:
- [ ] Go/No-Go meeting timestamp:
