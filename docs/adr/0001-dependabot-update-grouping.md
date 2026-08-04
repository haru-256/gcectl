# ADR-0001: Dependabot の更新 PR を minor/patch のみまとめる

- Status: Accepted
- Date: 2026-08-04

## Context

Dependabot は Go modules、GitHub Actions、Terraform の依存関係を月次で更新している。
Terraform ではプロバイダが複数ディレクトリに存在するため、グループ化しないとモジュール × プロバイダごとに個別 PR が増える。
一方で SemVer の major 更新は破壊的変更を含みやすく、まとめると原因切り分けが難しくなる。

## Decision

全エコシステム（gomod / github-actions / terraform）で次の方針とする。

- **minor / patch**: 1 本の PR にまとめる
- **major**: まとめず、依存ごとに個別 PR を出す

設定は `.github/dependabot.yml` の `groups` に `minor-patch` のみを定義し、`major` グループは定義しないことで実現する。
グループに含まれない major 更新は Dependabot が自動的に個別 PR として作成する。

Terraform については、対象プロバイダ（`hashicorp/google` / `hashicorp/google-beta`）を明示し、複数ディレクトリを横断して minor/patch を 1 PR に集約する。

## Consequences

### Positive

- minor / patch は挙動が変わらないことが多いため、まとめてレビュー・マージできる
- major は CI が落ちることが予想されるため、個別 PR にして失敗原因を特定しやすい
- Terraform のように同一プロバイダが複数ディレクトリに存在するケースでも、minor/patch の PR 数を抑えられる

### Negative

- major 更新が同時に複数ある場合は、個別 PR の数が増える
- Terraform の major 更新はディレクトリごとに個別 PR になり得る
