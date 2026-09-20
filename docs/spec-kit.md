# Spec Kit の使い方

このリポジトリでの仕様駆動開発（SDD）の進め方。アプリの起動方法は [README](../README.md) を参照。

[Spec Kit](https://github.com/github/spec-kit) を使い、**仕様 → 計画 → タスク → 実装**の順で進める。
仕様・計画・タスクが Markdown としてリポジトリに残るため、後から経緯を追える。

## 前提

`uv`（Python 3.11+）が必要。このリポジトリは **v1.0.8** で初期化してある。

```bash
uv tool install specify-cli==1.0.8
```

環境を確認する。`Claude Code (available)` と出れば使える。

```bash
specify check
```

## 基本の流れ

```text
/speckit-constitution  →  /speckit-specify  →  /speckit-plan  →  /speckit-tasks  →  /speckit-implement
（最初の 1 回だけ）
```

Claude Code でスラッシュコマンドとして実行する。ターミナルのコマンドではない。

## 何を実行すると何が生成されるか

| コマンド | 生成物 | 中身 |
| --- | --- | --- |
| `/speckit-constitution` | `.specify/memory/constitution.md` | プロジェクトの原則。最初の 1 回だけ。以降は原則が変わったときに更新する |
| `/speckit-specify <機能の説明>` | `specs/<連番>-<名前>/spec.md`<br>`specs/<連番>-<名前>/checklists/requirements.md` | 仕様と、その品質チェックリスト |
| `/speckit-plan <技術的な制約>` | `plan.md`・`research.md`<br>`data-model.md`・`contracts/`<br>`quickstart.md` | 実装計画、技術判断の記録、データモデル、API 契約、確認手順 |
| `/speckit-tasks` | `tasks.md` | 依存関係順に並べたタスク一覧 |
| `/speckit-implement` | （コード） | `tasks.md` に沿って実装する |

任意の品質ゲート: `/speckit-clarify`（仕様の曖昧さを詰める）、`/speckit-checklist`（要件の品質確認）、
`/speckit-analyze`（spec・plan・tasks の整合チェック）、`/speckit-converge`（実装の積み残し確認）。

## 置き場所

| パス | 中身 | Git |
| --- | --- | --- |
| `.specify/` | テンプレート、補助スクリプト、constitution | コミットする |
| `.claude/skills/speckit-*` | `/speckit-*` の中身 | コミットする |
| `specs/<連番>-<名前>/` | 機能ごとの仕様・計画・タスク | **コミットする**（レビュー対象） |
| `.specify/feature.json` | いま作業中の機能を指すポインタ | コミットしない（除外済み） |

## ルールの役割分担

同じルールを二重管理しないため、置き場所を分けている。

```text
CLAUDE.md          原典。方針そのもの
   ↓ 再表現
constitution       判断基準。/speckit-plan と /speckit-analyze が読む
   ↓ 委譲
.claude/rules/     手順レベルの細則（バックエンドの層分け、マイグレーションの書き方）
```

矛盾した場合は上が優先する。原則や技術スタックを変えるときは、`CLAUDE.md` と
`.specify/memory/constitution.md` を**同じコミットで**更新する。

## 書くときの約束

- **spec に実装方法を書かない。** エンドポイント設計・クエリ・層の分け方は `/speckit-plan` の領分
- **constitution には「すでに真であること」だけ書く。** 守れないルールを書くと plan と analyze がノイズを生む
- **最初の spec は小さい 1 機能にする。** 既存システム全体の仕様化を最初のテーマにしない

## 実例

`specs/001-toggle-todo-completion/` に、specify → plan → tasks を通した実物がある。
書き方に迷ったらこれを見る。

## このリポジトリ固有の注意

> [!IMPORTANT]
> 公式ドキュメントと実際の挙動が食い違う箇所がある。以下はこのリポジトリで確認した事実。

- **コマンドはハイフン区切り**（`/speckit-plan`）。公式ドキュメントの一部は `/speckit.plan` と
  ドット表記だが、それは GitHub Copilot などコマンド形式の統合向け。Claude Code のスキル統合では
  ハイフンになる（`.specify/integration.json` の `invoke_separator`）
- **仕様はリポジトリ直下の `specs/` に作られる。** `.specify/specs/` ではない
- **git 拡張は導入していない。** ブランチは従来どおり手で切る。Spec Kit がいまどの機能を見ているかは
  `.specify/feature.json` が決めるため、`git checkout` しても切り替わらない

## 初期化をやり直す場合

通常は不要。バージョンを上げるときなどに使う。

```bash
specify init --here --force --non-interactive --integration claude --script sh
```

> [!CAUTION]
> `--force` は管理対象パスのファイルを置き換えることがある。必ず作業をコミットしてから実行し、
> 実行後に `git diff` で差分を確認する。
