# Architecture (Hexagonal / Ports & Adapters)


**Goal:** Maintain a pure domain core, with application use-cases orchestrating work through interfaces (ports). All IO (HTTP, DB, crypto keys) lives in adapters.


## Layers
- **Domain**: `User`, `Session`, `Tenant` entities; no external imports
- **Application** (Use Cases): `SignUp`, `SignIn`, `Refresh`, `Me`
- **Adapters**: HTTP handlers, Postgres repositories, Ed25519/JWKS signer, Redis (future)