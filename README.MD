# Kleff Project Template

> **Build. Host. Scale. Your Way.** A base repository template for projects under the **Kleff**
> organization.

---

## 🧭 Overview

This repository provides the **standard Kleff structure**, **CI/CD policies**, and **legal
documentation** required for all organization projects.

It includes:

- 🔒 **Org-wide CI checks** via [`kleffio/.github`](https://github.com/kleffio/.github)
- 🧩 **Linting & formatting** configurations (YAML, Markdown)
- ⚖️ **Legal files:** `LICENSE`, `NOTICE`, and `SECURITY.md`
- 👥 **CODEOWNERS** and label definitions

Use this as the starting point when creating a new repository.

---

## 🚀 Usage

1. Click **Use this template** → create a new repo under [`kleffio`](https://github.com/kleffio).
2. Update the following:
   - Repository name and description
   - Project-specific details in this README
   - Add any required dependencies, workflows, or docs
3. Push changes and open a PR — the **Org Policies** workflow will automatically run quality checks.

---

## ♻️ Org Workflows

This repository uses the centralized **Org Policies** workflow from `kleffio/.github`.

```yaml
name: Kleff Org Policies
on:
  pull_request:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  org-policies:
    uses: kleffio/.github/.github/workflows/_org-policies.yml@main
```

Checks include:

- YAML linting
- Markdown linting
- Actionlint
- PR title & required labels
- OpenSSF Scorecard

---

## 🧩 Contributing

Follow the
[Kleff Contribution Guidelines](https://github.com/kleffio/.github/blob/main/CONTRIBUTING.md). Open
PRs against `main` unless noted otherwise.

---

## 🔒 Security

For vulnerability reports, see [`SECURITY.md`](./SECURITY.md).

---

## ⚖️ License

Licensed under the **Kleff Public License v1.0**. See [`LICENSE`](./LICENSE) and
[`NOTICE`](./NOTICE) for details.

---

> © 2025 Kleff Hosting — Build. Host. Scale. Your Way.
