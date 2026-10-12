# Specification Quality Checklist: 顧客のアカウント作成・ログイン・ログアウト（自前認証）

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-12
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notes

- Items marked incomplete require spec updates before `/speckit-clarify` or `/speckit-plan`
- 検証 1 回目で、用語の不一致を 1 件直した（「会員情報」→「お客様の情報」。`01` の 4 章で「会員」は使ってはいけない別名）。
- ログアウトの異常系を、`04` の `DELETE /sessions`（未認証は 401）と矛盾しない書き方に直した。
- [NEEDS CLARIFICATION] は 0 件。Issue #13 の「決めること」は次のように既定を置き、Assumptions に明記した。
  - 確認メール：送信の口だけ用意し、送らない
  - 認証主体（`auth_issuer` / `auth_subject`）：`03` に合わせて空のまま
  - 認証情報の 1:N 化：本機能では決めず、設計書どおり 1:1
- そのほか、セッションの有効期限（固定 30 日）、パスワードの長さ（8 文字以上 72 バイト以下）、
  メールアドレスの大文字・小文字の扱いは業界の一般的な既定を置いた。変えたい場合は `/speckit-clarify` で調整する。
- 利用者の指示で、アカウント作成をメールアドレスとパスワードだけにした。氏名・フリガナ・電話番号の登録はプロフィール編集（別機能）に分け、
  `customers` の該当列を未登録可にする設計書の変更を Assumptions に明記した。電話番号とフリガナの入力条件は FR-002 から外した。
- FR-020（将来の切り替えへの備え）は利用者から直接は見えない。外部の認証方式に差し替えても FR-014〜FR-016 の
  受け入れ条件が変わらないことで検証する。
- Assumptions に設計書のファイル名を出しているのは、設計書との対応を追えるようにするためで、実装の指示ではない。
