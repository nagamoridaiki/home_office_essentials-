# EC MVP API 一覧

作成日: 2026-10-09 / 対象: Go 標準 `net/http`（Go 1.22+ のメソッド付きパターン）

## 0. このドキュメントについて

- **正はこのリポジトリ**。Drive のスプレッドシートは人が読み返すためのスナップショット。
- 用語は [用語集](01-ubiquitous-language.md)、ステータスは [状態遷移表](02-state-transitions.md)、テーブルは [論理データモデル](03-logical-data-model.md) を正とする。
- **「走るタイミング」はどの画面のどの操作で呼ばれるか**、**「副作用」はレスポンス以外に起きる変化**を書いた。実装時に見落としやすいのはこの 2 列。

## 1. 共通規約

| 項目 | 決めたこと |
| --- | --- |
| ベース URL | 開発は `http://localhost:8080`。**パスにプレフィックスは付けない** |
| パス | `kebab-case` の複数形。`/api` もバージョン番号も付けない |
| 顧客自身のリソース | `/me/` 配下に寄せる。**`/me/` 配下はセッションの顧客に絞る**ことをミドルウェアで一括強制し、ハンドラごとの所有者チェックを書かない |
| 管理者 | `/admin/` 配下。管理者セッションだけが通る |
| 識別子 | 顧客向けは `public_id`（uuid）、管理者向けは `id`（bigint）。顧客向けに連番を出さない |
| JSON キー | `snake_case`。DB のカラム名と揃える |
| 金額 | 円単位の整数。小数・文字列にしない |
| 日時 | RFC 3339（`2026-10-09T12:00:00+09:00`） |
| 認証 | セッション Cookie（`HttpOnly` / `Secure` / `SameSite=Lax`）。DB にはトークンのハッシュだけを保存する |
| CORS | 開発はフロントのオリジン（`http://localhost:3000`）に限定する |
| エラー形式 | `{"error":{"code":"...","message":"...","details":[...]}}`。`code` は機械判定用の固定文字列、`message` は日本語 |
| ステータスコード | 400 入力不正 / 401 未認証 / 403 権限なし / 404 なし / **409 状態の競合**（在庫切れ・ロック期限切れ・遷移不正）/ 422 業務ルール違反 / 500 |
| 一覧のページング | `cursor` + `limit`。`offset` は使わない（件数が増えると重くなり、途中で挿入されるとずれる） |
| 冪等性 | Webhook はイベント ID で重複を弾く。顧客の POST は二重送信でも同じ結果を返し、重ねて作らない。ただし `POST /customers` は例外で、2 回目の同じ要求は `409 email_already_used` を返す（アカウントが 2 つ作られない点は同じ） |

## 2. エンドポイント一覧（59 本）

### 2.1 公開（認証不要・5 本）

| メソッド | パス | 用途 | 走るタイミング | 認証 | リクエスト概要 | レスポンス概要 | 副作用 | 主なエラー |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| GET | `/categories` | カテゴリの一覧を返す | 商品一覧の初回表示、検索メニューを開いたとき | 不要 | — | `items[]`（`id`, `name`, `slug`, `parent_id`） | なし | — |
| GET | `/selling-products` | 販売枠を検索する | 商品一覧の表示、絞り込み操作、追加読み込み | 不要 | `q`, `category`, `condition_rank`, `min_price`, `max_price`, `cursor`, `limit` | `items[]`（名前・価格・ランク・在庫残数・代表画像）, `next_cursor` | なし | 400 条件不正 |
| GET | `/selling-products/{id}` | 販売枠の詳細を返す | 商品詳細ページを開いたとき | 不要 | — | 販売枠、画像一覧、在庫残数、コンディション詳細、配送目安 | なし | 404 |
| POST | `/customers` | アカウントを作る | 新規登録の最終ステップで送信したとき | 不要 | `email`, `password` | 自分の情報 | `customers` と `customer_credentials` を作る。確認メールを送る。セッション Cookie を発行する | 409 メール重複、422 検証失敗 |
| POST | `/sessions` | ログインする | ログイン画面で送信したとき | 不要 | `email`, `password` | 自分の情報 | `customer_sessions` を作り Cookie を発行する | 401 認証失敗 |

### 2.2 顧客（顧客セッション・23 本）

| メソッド | パス | 用途 | 走るタイミング | 認証 | リクエスト概要 | レスポンス概要 | 副作用 | 主なエラー |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| DELETE | `/sessions` | ログアウトする | ヘッダーのログアウトを押したとき | 顧客 | — | 204 | `customer_sessions.revoked_at` を埋め、Cookie を失効させる | 401 |
| GET | `/me` | 自分の情報を返す | マイページを開いたとき、ヘッダーの表示 | 顧客 | — | `email`、氏名・フリガナ・`phone_number`（未登録なら `null`） | なし | 401 |
| PATCH | `/me` | 自分の情報を更新する | プロフィール編集を保存したとき。氏名・フリガナ・電話番号の初回登録もここで行う | 顧客 | 変更する項目だけ | 更新後の自分の情報 | `customers` を更新する | 422 検証失敗 |
| GET | `/me/addresses` | 住所帳を返す | 配送先選択、マイページの住所管理 | 顧客 | — | `items[]`（住所と `is_default`） | なし | 401 |
| POST | `/me/addresses` | 住所を追加する | 配送先入力で新規住所を保存したとき | 顧客 | 宛名・郵便番号・都道府県・市区町村・番地・建物・電話・`is_default` | 作った住所 | `is_default` が true なら他の住所の既定を外す | 422 検証失敗 |
| PATCH | `/me/addresses/{id}` | 住所を更新する | 住所の編集を保存したとき | 顧客 | 変更する項目だけ | 更新後の住所 | 同上 | 404, 422 |
| DELETE | `/me/addresses/{id}` | 住所を削除する | 住所管理で削除したとき | 顧客 | — | 204 | 行を消す。過去の注文の届け先は `shipping_addresses` に固定済みなので変わらない | 404 |
| GET | `/me/bank-accounts` | 振込先口座の一覧を返す | 査定結果確認画面を開いたとき | 顧客 | — | `items[]`（金融機関・支店・種別・口座番号・名義） | なし | 401 |
| POST | `/me/bank-accounts` | 振込先口座を登録する | 査定結果確認画面で口座を入力したとき | 顧客 | 金融機関・支店・`account_type`・口座番号・名義カナ | 作った口座 | `bank_accounts` を作る | 422 検証失敗 |
| POST | `/me/checkout-sessions` | **購入手続きを開始し、在庫を 15 分ソフトロックする** | 商品詳細で「購入手続きに進む」を押した瞬間 | 顧客 | `items[].selling_product_id`, `quantity` | `id`, `expires_at`, 確保した在庫の内容 | **空いた在庫を 1 点選んで `locked` にし、`lock_checkout_session_id` と `lock_expires_at` を埋める。`inventory_movements` に `locked` を残す** | **409 在庫切れ（空いた個体がない）**、401 |
| GET | `/me/checkout-sessions/{id}` | 手続きの内容と残り時間を返す | 配送・決済画面の表示、タイマーの再同期 | 顧客 | — | `expires_at`, `items[]`, 金額の内訳 | なし | 404, 409 期限切れ |
| PATCH | `/me/checkout-sessions/{id}` | 配送先と希望配送日を設定する | 配送先・お届け日を選んだとき | 顧客 | `address_id`, `requested_delivery_date`, `requested_time_slot` | 更新後の金額の内訳 | 配送料と消費税を再計算する | 404, 409 期限切れ |
| DELETE | `/me/checkout-sessions/{id}` | 手続きを中断してロックを解放する | 「戻る」を押したとき、タイムアウトモーダルを閉じたとき | 顧客 | — | 204 | **lock の 2 列を NULL に戻し、在庫を `available` にする。`inventory_movements` に `lock_released` を残す** | 404 |
| POST | `/me/checkout-sessions/{id}/payment-intent` | Stripe の PaymentIntent を作り `client_secret` を返す | 決済フォームを表示する直前 | 顧客 | — | `client_secret`, `amount` | Stripe に PaymentIntent を作り、`checkout_sessions.stripe_payment_intent_id` を埋める | **409 ロック期限切れ**、404 |
| GET | `/me/selling-orders` | 自分の注文履歴を返す | マイページの注文履歴を開いたとき | 顧客 | `cursor`, `limit` | `items[]`（`public_id`, `order_number`, 状態, 金額, 代表商品）, `next_cursor` | なし | 401 |
| GET | `/me/selling-orders/{public_id}` | 注文の詳細を返す | 注文履歴から詳細を開いたとき | 顧客 | — | 明細・金額内訳・届け先・配送状況・決済状況 | なし | 404 |
| POST | `/me/purchase-orders` | 買取を申し込む | 買取申込フォームを送信したとき | 顧客 | `declared_model_number`, `declared_rank`, `declared_note` | `public_id`, `order_number`, 状態 | `purchase_orders` を `requested` で作り、`BUY-` の番号を採番する | 422 検証失敗 |
| POST | `/me/purchase-orders/{public_id}/photos` | 申込写真をアップロードする | 申込フォームで写真を追加したとき | 顧客 | 画像ファイル、`label`, `customer_note` | 作った写真の情報 | 画像ストレージに保存し、`purchase_order_photos` を作る | 409 状態が `requested` でない、413 サイズ超過 |
| PATCH | `/me/purchase-orders/{public_id}/pickup` | 集荷希望日と時間帯を登録する | 集荷希望日を選んで送信したとき | 顧客 | `pickup_scheduled_date`, `pickup_time_slot` | 確定した集荷情報と `pickup_fee` | 時間帯指定に応じて `pickup_fee` を確定する | 409 状態が `identity_verified` でない |
| GET | `/me/purchase-orders` | 自分の買取依頼の一覧を返す | マイページの買取履歴を開いたとき | 顧客 | `cursor`, `limit` | `items[]`（`public_id`, `order_number`, 状態, 査定額）, `next_cursor` | なし | 401 |
| GET | `/me/purchase-orders/{public_id}` | 買取依頼の詳細と査定結果を返す | 査定結果確認画面を開いたとき | 顧客 | — | 申告内容・提出写真・暫定査定額・最終査定額・減額明細・回答期限 | なし | 404 |
| POST | `/me/purchase-orders/{public_id}/agreement` | **最終査定額に合意する** | 査定結果確認画面で「合意する」を押したとき | 顧客 | `bank_account_id`, `agreed_terms` | 状態, `agreed_at`, 送金予定額 | **`agreed_at` を埋め、在庫を `preparing` で 1 件作る。`inventory_movements` に `receiving` を残す** | 409 状態が `inspected` でない、422 口座未登録 |
| POST | `/me/purchase-orders/{public_id}/return` | 返品を依頼する | 「返品を依頼」の確認モーダルで確定したとき | 顧客 | — | 状態, `returned_at` | `returned_at` を埋め、返送を手配する。返送料は請求しない | 409 状態が `inspected` でない |

### 2.3 管理者（管理者セッション・30 本）

| メソッド | パス | 用途 | 走るタイミング | 認証 | リクエスト概要 | レスポンス概要 | 副作用 | 主なエラー |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| POST | `/admin/sessions` | 管理者としてログインする | 管理画面のログイン送信時 | 不要 | `email`, `password` | 管理者の情報 | `admin_sessions` を作り Cookie を発行する | 401 認証失敗 |
| DELETE | `/admin/sessions` | 管理者としてログアウトする | サイドバーのログアウト | 管理者 | — | 204 | `admin_sessions.revoked_at` を埋める | 401 |
| GET | `/admin/dashboard` | 全体のサマリーを返す | 管理画面のダッシュボードを開いたとき | 管理者 | — | 在庫サマリー、要対応件数、ロック中の件数 | なし | 401 |
| GET | `/admin/selling-products` | 販売枠の一覧を返す | 販売枠の管理画面を開いたとき | 管理者 | `q`, `category`, `listing_status`, `cursor` | `items[]`, `next_cursor` | なし | 401 |
| POST | `/admin/selling-products` | 販売枠を作る | 新規登録フォームを保存したとき | 管理者 | 型番・名前・区分・ランク・価格・説明・カテゴリ | 作った販売枠 | `selling_products` を `draft` で作る | 409 型番とランクの重複、422 |
| GET | `/admin/selling-products/{id}` | 販売枠の詳細を返す | 販売枠の編集画面を開いたとき | 管理者 | — | 販売枠、画像、紐づく在庫の件数 | なし | 404 |
| PATCH | `/admin/selling-products/{id}` | 販売枠を更新する | 編集を保存したとき、掲載状態を切り替えたとき | 管理者 | 変更する項目だけ | 更新後の販売枠 | 価格を変えても過去の注文明細の単価は変わらない | 404, 422 |
| POST | `/admin/selling-products/{id}/images` | 商品画像を追加する | 画像をアップロードしたとき | 管理者 | 画像ファイル、`display_order`, `alt_text` | 作った画像の情報 | 画像ストレージに保存する | 409 表示順の重複 |
| DELETE | `/admin/selling-products/{id}/images/{image_id}` | 商品画像を削除する | 画像の削除ボタン | 管理者 | — | 204 | 行を消す。ストレージの実体も消す | 404 |
| GET | `/admin/inventories` | 在庫の一覧を返す | 在庫一覧を開いたとき、要対応で絞り込んだとき | 管理者 | `inventory_status`, `model_number`, `condition_rank`, `cursor` | `items[]`（状態・ロック残り時間・棚番・原価）, `next_cursor` | なし | 401 |
| POST | `/admin/inventories` | 新品を入庫登録する | 仕入れを登録したとき | 管理者 | `selling_product_id`, `cost_price`, `storage_location` | 作った在庫 | 在庫を `preparing` で作り、`inventory_movements` に `receiving` を残す | 422 |
| GET | `/admin/inventories/{id}` | 在庫 1 点の詳細を返す | 在庫詳細を開いたとき | 管理者 | — | 状態・ロック情報・原価・棚番・入庫元・出庫先 | なし | 404 |
| PATCH | `/admin/inventories/{id}` | 棚番や原価を更新する | 在庫詳細で保存したとき | 管理者 | `storage_location`, `cost_price` | 更新後の在庫 | `inventory_movements` に記録を残す | 404, 409 `sold` は編集不可 |
| POST | `/admin/inventories/{id}/listing` | 出品する（販売枠に紐づけて販売可にする） | 在庫詳細で「出品する」を押したとき | 管理者 | `selling_product_id` | 更新後の在庫 | 在庫を `available` にし、`inventory_movements` に `listed` を残す | 409 状態が `preparing` / `suspended` でない |
| DELETE | `/admin/inventories/{id}/listing` | 出品を停止する | 「出品を停止」を押したとき | 管理者 | — | 更新後の在庫 | 在庫を `suspended` にし、`unlisted` を残す | 409 ロック中は停止できない |
| POST | `/admin/inventories/{id}/discard` | 廃棄する | 危険操作エリアで廃棄を確定したとき | 管理者 | `discard_reason` | 更新後の在庫 | 在庫を `discarded` にし、`discarded` を残す | 409 状態が `preparing` / `suspended` でない |
| GET | `/admin/inventories/{id}/movements` | 在庫変動履歴を返す | 在庫詳細の履歴セクション | 管理者 | `cursor` | `items[]`（日時・種別・関連取引・実行者）, `next_cursor` | なし | 404 |
| GET | `/admin/selling-orders` | 販売注文の一覧を返す | 販売注文一覧を開いたとき、タブで絞ったとき | 管理者 | `selling_order_status`, `q`, `cursor` | `items[]`（番号・顧客・金額・決済状態・配送状態）, `next_cursor` | なし | 401 |
| GET | `/admin/selling-orders/{id}` | 販売注文の詳細を返す | 販売注文詳細を開いたとき | 管理者 | — | 明細・金額内訳・届け先・決済・配送・顧客情報 | なし | 404 |
| POST | `/admin/selling-orders/{id}/shipment` | 送り状を発行し発送済にする | 配送セクションで「発行する」を押したとき | 管理者 | `carrier_code`, `tracking_number`, お届け日時 | 作った配送情報 | 注文を `shipped` にし、`shipments` を作る。在庫の `sold_at` は変えない | 409 状態が `confirmed` でない |
| PATCH | `/admin/selling-orders/{id}/shipment` | 配送完了を記録する、または送り状を再発行する | 「配送完了にする」、業者や日時を変えて再発行したとき | 管理者 | `delivered_at` または `carrier_code` と `tracking_number` | 更新後の配送情報 | 注文を `delivered` にする | 409 状態が `shipped` でない |
| POST | `/admin/selling-orders/{id}/cancellation` | 発送前の注文を取り消す | 注文詳細でキャンセルを確定したとき | 管理者 | `cancel_reason` | 更新後の注文 | 注文を `canceled` にし、Stripe で返金する。**在庫は `sold` のまま**（同じ個体を売るときは新しい在庫行を作る） | 409 発送後は不可 |
| POST | `/admin/selling-orders/{id}/refund` | 返金する | 返品を受け付けて返金したとき | 管理者 | `amount`, `reason` | 更新後の注文と決済 | 注文を `refunded` にし、Stripe で返金する。`payments.refunded_at` を埋める | 409 状態が不正、422 金額超過 |
| GET | `/admin/purchase-orders` | 買取依頼の一覧を返す | 買取依頼一覧を開いたとき、要対応で絞ったとき | 管理者 | `purchase_order_status`, `q`, `cursor` | `items[]`（番号・顧客・申告品・査定進捗・物流・査定額）, `next_cursor` | なし | 401 |
| GET | `/admin/purchase-orders/{id}` | 買取依頼の詳細を返す | 買取注文詳細を開いたとき | 管理者 | — | 申告内容・写真・本人確認・物流・査定・振込先 | なし | 404 |
| POST | `/admin/purchase-orders/{id}/photo-assessment` | 写真査定を確定し暫定査定額を出す | 査定ワークベンチで写真査定を確定したとき | 管理者 | `estimated_amount`, 写真ごとの評価 | 更新後の買取注文 | 状態を `photo_assessed` にし、`estimated_amount` を埋める。顧客に通知する | 409 状態が `requested` でない |
| POST | `/admin/purchase-orders/{id}/identity-verification` | 本人確認の結果を記録する | eKYC パネルでチェックを保存したとき | 管理者 | `name_matched`, `address_matched` | 本人確認の記録 | `identity_verifications` を作り、両方 true なら状態を `identity_verified` にする | 409 状態が `photo_assessed` でない |
| POST | `/admin/purchase-orders/{id}/arrival` | 倉庫到着を記録する | 荷物の受け取りを登録したとき | 管理者 | `arrived_at` | 更新後の買取注文 | 状態を `arrived` にし、`arrived_at` を埋める | 409 状態が `pickup_scheduled` でない |
| POST | `/admin/purchase-orders/{id}/inspection` | 実物検品を確定し、減額を反映して最終査定額を出す | 査定ワークベンチで「確定して承認依頼を送る」を押したとき | 管理者 | `condition_rank`, `deductions[]`（`photo_id`, `reason`, `amount`） | `final_assessed_amount`, `agreement_deadline_at` | 状態を `inspected` にし、`purchase_order_deductions` を作る。回答期限を設定し、顧客に通知する | 409 状態が `arrived` でない、422 減額が暫定額を超える |
| POST | `/admin/purchase-orders/{id}/payout` | 合意済みの買取金を送金する | 「送金する」を押したとき | 管理者 | — | `payouts` の記録 | 状態を `paid` にし、Stripe で送金する。送金額は `final_assessed_amount − pickup_fee` | **409 状態が `agreed` でない**、422 口座不正 |

### 2.4 Webhook（1 本）

| メソッド | パス | 用途 | 走るタイミング | 認証 | リクエスト概要 | レスポンス概要 | 副作用 | 主なエラー |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| POST | `/webhooks/stripe` | **注文確定の正。**入金を受けて販売注文を作る | 顧客が決済を完了した後、Stripe から非同期で届く（再送あり） | Stripe 署名 | Stripe のイベント本体（`payment_intent.succeeded` など） | 200（処理済みのイベントでも 200） | **`stripe_webhook_events` に記録し、同一トランザクション内で 在庫を `SELECT ... FOR UPDATE` → 注文と明細を作成 → 在庫を `sold` → lock の 2 列を NULL → `payments` を作成 → `inventory_movements` に `sold`** | 400 署名不正。処理失敗は 500 を返して Stripe に再送させる |

## 3. 不変条件に関わる 3 本

この 3 本だけは、**何を何順でやるか**が不変条件そのものになる。

### 3.1 `POST /me/checkout-sessions` — ロックを取る

1. トランザクション開始
2. 対象の販売枠から、空いた在庫を 1 行だけ掴む
   `SELECT id FROM inventories WHERE selling_product_id = $1 AND inventory_status = 'available' AND (lock_expires_at IS NULL OR lock_expires_at <= now()) ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 1`
3. 0 件なら 409 を返して終わり（在庫切れ）
4. `checkout_sessions` を作る
5. 在庫を `locked` にし、`lock_checkout_session_id` と `lock_expires_at = now() + interval '15 minutes'` を埋める
6. `inventory_movements` に `locked` を残す
7. コミット

`FOR UPDATE SKIP LOCKED` が要。同時に来た 2 人目は 1 人目が掴んだ行を**待たずに飛ばして**次の個体を取るので、どちらも待たされずに別々の在庫を確保できる。

### 3.2 `POST /webhooks/stripe` — 注文を確定する

1. 署名を検証する（失敗なら 400）
2. `stripe_webhook_events` に `stripe_event_id` で INSERT を試みる。重複なら**何もせず 200**（冪等化）
3. トランザクション開始
4. `payment_intent` から `checkout_sessions` を引き、その在庫を `SELECT ... FOR UPDATE` で掴む
5. ロックが期限内で、在庫が `locked` であることを確認する（違えば記録を残して 200。入金済みなので返金の対象にする）
6. `selling_orders` と `selling_order_items` を作る。`order_number` を採番する
7. 在庫を `sold` にし、`sold_at` を埋め、lock の 2 列を NULL に戻す
8. `payments` を作る
9. `inventory_movements` に `sold` を残す
10. `checkout_sessions` を `completed` にする
11. コミットして 200

`selling_order_items.inventory_id` の UNIQUE が最後の砦。万一 2 回通っても 2 件目は制約違反で落ちる。

### 3.3 `POST /me/purchase-orders/{public_id}/agreement` — 合意して在庫にする

1. トランザクション開始
2. 買取注文を `SELECT ... FOR UPDATE` で掴み、状態が `inspected` であることを確認する（違えば 409）
3. 回答期限を過ぎていないことを確認する（過ぎていれば 409。期限超過は自動返送の対象）
4. `bank_account_id` が自分の口座であることを確認する（違えば 422）
5. `agreed_at` と `bank_account_id` を埋め、状態を `agreed` にする
6. **`inventories` に 1 件作る**（`purchase_order_id` を入れ、`cost_price` に `final_assessed_amount`、状態は `preparing`）
7. `inventory_movements` に `receiving` を残す
8. コミット

ここで在庫ができるので、所有権が移った瞬間と帳簿が一致する。送金（`POST /admin/purchase-orders/{id}/payout`）は別の操作で、失敗しても在庫は残る。

> 補足: この 3 本のシーケンス図は未作成（依頼待ち）。
