# EC MVP 用語集（ユビキタス言語）

作成日: 2026-09-20 / 対象: home_office_essentials MVP（新品・中古の EC 販売 ＋ 買取の循環）

## 0. このドキュメントについて

- ここで決めた名前を、日本語の会話・Go のコード・DB のカラム・API・UI 文言のすべてで揃える。勝手な別名を作らない。
- **正はこのリポジトリ（`docs/design/`）**。Google ドライブに置くものは人が読み返すためのスナップショットで、内容が食い違ったらリポジトリが正しい。
- 参考にした既存資料（Drive の要件定義・データ設計・API 一覧・画面設計）とは、意図的に名前を変えた箇所がある。理由は「使ってはいけない別名」欄に書いた。

### Drive 上のスナップショット

同じ内容を、人が読み返しやすいように Drive にも置いてある。内容が食い違ったらこのファイルが正しい。

| 形式 | ファイル | 収録範囲 |
| --- | --- | --- |
| ドキュメント | [EC MVP 用語集（ユビキタス言語）](https://docs.google.com/document/d/1JWIWJdXH8gHRjitccGPKUaqkho3WXb4L8gmxuhPPihw/edit) | このファイル全体 |
| スプレッドシート | [用語一覧](https://docs.google.com/spreadsheets/d/1TPLw6nDgE3HnThh0pZfE2vskSmZG0T33gnwbKLBt6KA/edit) | 3 章（区分・MVP 列を追加して 1 枚に統合） |
| スプレッドシート | [禁止別名早見表](https://docs.google.com/spreadsheets/d/1xyfewtu6fItpi6hckMzUs3Ln2MMLRPULlCVoIFVZXZI/edit) | 4 章 |
| スプレッドシート | [命名規約](https://docs.google.com/spreadsheets/d/19MyX_2ivT1ScLsyQpNS7ijwJTIfYMyE3FSFh4CmKv2I/edit) | 2 章 |

## 1. 前提と未決事項

### 1.1 MVP のスコープ

買い取ったものを検品して中古として再出品するところまでで一周が閉じる。この循環に関係しない機能は用語だけ予約し、定義を作り込まない（3.7 節）。

### 1.2 置いた前提

| 論点 | 置いた前提 | 理由 / 後で変えるときの影響 |
| --- | --- | --- |
| 認証 | MVP は自前実装（メールアドレス＋パスワード、セッション Cookie） | 将来 Amazon Cognito / Google Cloud Identity Platform などの **OIDC に切り替える予定**。そのためドメイン層は顧客 ID しか知らず、パスワードと認証方式はインフラ層に閉じる。顧客は `auth_subject`（＋ `issuer`）で認証結果と結びつけ、OIDC 移行時はここに IdP の `sub` と `iss` を入れるだけで済むようにする |
| 決済手段 | MVP は **カードのみ（Stripe）** | 15 分の在庫ソフトロックと、支払期限が数日ある銀行振込・コンビニ払い・代引きは両立しない。既存の画面設計にある 5 決済手段は MVP 対象外 |

### 1.3 用語が守る不変条件

| 不変条件 | 関係する用語 |
| --- | --- |
| 1 点ものが二重に売れない | 在庫 / 在庫ソフトロック / ロック期限 / 販売枠 |
| 在庫ステータスが不正に遷移しない | 在庫ステータス / 販売注文ステータス / 買取注文ステータス |
| 他人の個人データ・税務データに URL を書き換えても到達できない | 顧客 / 購入者 / 売却者 / 認証主体 / 管理者 |

## 2. 命名規約

| 対象 | 規約 | 例 |
| --- | --- | --- |
| 日本語の用語 | この用語集の「日本語の用語」欄の表記に統一する | 販売枠 |
| Go の型・関数 | `PascalCase`。単数形 | `SellingProduct` |
| DB のテーブル | `snake_case`。複数形 | `selling_products` |
| DB のカラム | `snake_case`。単数形 | `selling_product_id` |
| API のパス | `/api/` 配下。`kebab-case` の複数形。バージョン番号は付けない | `GET /api/selling-products/{id}` |
| 管理者向け API | `/api/admin/` 配下に分ける | `GET /api/admin/purchase-orders/{id}` |
| JSON のキー | `snake_case`。DB のカラム名と揃える | `"final_assessed_amount"` |
| UI 文言 | 顧客向けは平易な日本語。コード上の名前と違ってよいが、この用語集に必ず書く | 「買取依頼」 |
| 金額 | 円を整数（`int64`）で持つ。小数・浮動小数点は使わない | `selling_price` |
| 日時 | `timestamptz`。カラム名は `_at` で終える | `expires_at` |
| 監査列 | 全テーブルに `created_at` / `updated_at`。`updated_at` は DB トリガーで更新（`.claude/rules/database.md`） | — |
| 主キー | `bigint GENERATED ALWAYS AS IDENTITY`。文字列の主キーは使わない | `id` |

## 3. 用語集

### 3.1 アクターと共通概念

| 日本語の用語 | コード上の名前 | API 上の名前 | UI 文言 | 定義 | 使ってはいけない別名 |
| --- | --- | --- | --- | --- | --- |
| 顧客 | `Customer` / `customers` | `/api/customers`, `customer_id` | お客様 | 本サービスに登録されたアカウント 1 件。買う人と売る人を別のアカウントには分けず、同じ顧客が両方を行う | ユーザー、会員、カスタマー、`users` |
| 購入者 | `Buyer` | `buyer_id` | （画面に出さない） | 販売注文から見た顧客の立場。顧客の属性ではなく役割なので、顧客テーブルに購入者フラグは持たせない | 購入ユーザー、バイヤー、お客さん |
| 売却者 | `Seller` | `seller_id` | お客様 | 買取注文から見た顧客の立場。運営が商品を売り出すこととは無関係 | 出品者、売主、ベンダー、サプライヤー |
| 管理者 | `Admin` | `/api/admin/...` | 管理画面 | 運営側の操作者。MVP では運営者 1 人だけを想定し、権限の細分化はしない | スタッフ、オペレーター、店舗、運営者（コード・API では不可） |
| 認証主体 | `AuthSubject` | `auth_subject`, `issuer` | （画面に出さない） | 顧客を認証結果と結びつける識別子。MVP は自前で発行し、OIDC 移行時は IdP の `sub` と `iss` を入れる | ログイン ID、アカウント ID、`uid` |
| 金額 | `Amount` (`int64`) | `*_amount`, `*_price` | 円 | 日本円を円単位の整数で表したもの。丸め誤差を持ち込まないため小数は使わない | 価格（文脈を限定しない使い方）、`price`（単独での汎用利用） |

### 3.2 カタログ（販売枠）

| 日本語の用語 | コード上の名前 | API 上の名前 | UI 文言 | 定義 | 使ってはいけない別名 |
| --- | --- | --- | --- | --- | --- |
| 販売枠 | `SellingProduct` / `selling_products` | `/api/selling-products`, `selling_product_id` | 商品 | 同じ型番・同じコンディションランクの個体をまとめて売り出す単位。価格・説明文・画像はここが持つ。実在庫とは 1:N | 商品マスター、SKU、プロダクト、商品（コード・API では不可） |
| 型番 | `ModelNumber` | `model_number` | 型番 | メーカーが定めた機種の識別子。販売枠と在庫をグルーピングする軸になる | 品番、モデル名、`product_code` |
| 商品区分 | `ProductCondition` | `product_condition` | 新品 / 中古 | その販売枠が新品か中古かの区分。値は `new` と `used` の 2 つだけ | 新中区分、`is_used`、`is_new` |
| コンディションランク | `ConditionRank` | `condition_rank` | コンディション | 中古個体の状態を表す等級。値は `S` `A` `B_PLUS` `B` `C` `D` の 6 段階。新品には付けない | グレード、ランク（単独）、状態、`rank`（単独） |
| 商品画像 | `SellingProductImage` / `selling_product_images` | `/api/selling-products/{id}/images`, `image_url` | 商品写真 | 販売枠に紐づく画像。枚数可変の別テーブルとして持つ。既存資料の 10 カラム方式は使わない | `product_images_1` 〜 `product_images_10`、サムネ |
| カテゴリ | `Category` / `categories` | `/api/categories`, `category_id` | カテゴリ | 家具・家電・周辺機器といった商品の分類 | ジャンル、種別、タグ |
| 販売価格 | `SellingPrice` | `selling_price` | 価格（税込） | 販売枠に設定した、顧客に提示する 1 点あたりの税込価格 | 売価、単価、`amount`（単独） |
| 商品説明 | `Description` | `description` | 商品説明 | 運営が書く販売枠の説明文。買取時に顧客が書いた申告メモとは別物 | `admin_description`、`seller_description`、備考 |
| 掲載状態 | `ListingStatus` | `listing_status` | 掲載中 / 掲載停止 | 販売枠を顧客に見せるかどうか。値は `draft` `published` `suspended`。個体の在庫ステータスとは別軸 | 公開フラグ、`is_published`、出品ステータス |

### 3.3 在庫

| 日本語の用語 | コード上の名前 | API 上の名前 | UI 文言 | 定義 | 使ってはいけない別名 |
| --- | --- | --- | --- | --- | --- |
| 在庫 | `Inventory` / `inventories` | `/api/admin/inventories`, `inventory_id` | 在庫 | 実在する個体 1 点。1 点ものなので数量を持たず、1 行がちょうど 1 個体を表す。買取由来か新品仕入れかを問わない | 在庫数、ストック、商品、買取商品、`purchased_product` |
| 在庫ステータス | `InventoryStatus` | `inventory_status` | ステータス | 個体が今どの工程にいるか。許可された遷移だけを通す（状態遷移表を参照）。二重販売を防ぐ要の値 | 在庫状態、ステート、`status`（単独） |
| 在庫ソフトロック | `InventoryLock` / `inventory_locks` | `POST /api/checkout-sessions` で発行 | 在庫を確保しています | 「購入手続きに進む」を押した時点で、特定の在庫 1 点を 15 分間だけ予約する記録。Postgres だけで完結させ、Redis は使わない。期限切れの行は削除せず失効として残す | 仮押さえ（UI 文言としてのみ可）、カート確保、予約、`is_rock`、`is_temporary` |
| ロック期限 | `ExpiresAt` | `expires_at` | 残り mm:ss | ソフトロックが失効する時刻。発行時刻の 15 分後。これを過ぎたロックは在庫の確保とみなさない | タイムアウト、有効期限、`limit_at` |
| 入庫 | `Receiving` | `inventory_movements.type = "receiving"` | 入庫 | 個体を在庫として登録すること。買取の検品完了と、新品の仕入れの 2 経路がある | 仕入、登録、`stock_in` |
| 出品 | `Listing` / `ListInventory` | `POST /api/admin/inventories/{id}/listing` | 出品する | 在庫を販売枠に結びつけて、顧客が買える状態にすること | 公開、掲載開始、アップ |
| 出品停止 | `Unlisting` | `DELETE /api/admin/inventories/{id}/listing` | 出品を停止 | 在庫を販売対象から外すこと。在庫そのものは消さない | 削除、取り下げ、非公開化 |
| 保管場所 | `StorageLocation` | `storage_location` | 棚番号 | 個体を置いている倉庫の棚。ピッキング指示に使う | `saved_place`、置き場、ロケーション |
| 仕入原価 | `CostPrice` | `cost_price` | 原価 | その個体を手に入れるのにかかった金額。買取なら最終査定額、新品なら仕入価格 | 買値、`purchased_price`、コスト |
| 在庫変動履歴 | `InventoryMovement` / `inventory_movements` | `/api/admin/inventories/{id}/movements` | 在庫変動履歴 | 入庫・出品・ロック・販売といった在庫ステータスの変化を時系列で残す記録。監査証跡になる | 在庫ログ、履歴（単独）、`inventory_history` |

### 3.4 販売

| 日本語の用語 | コード上の名前 | API 上の名前 | UI 文言 | 定義 | 使ってはいけない別名 |
| --- | --- | --- | --- | --- | --- |
| カート | `Cart` | `/api/cart` | カート | 顧客が検討中の販売枠を入れておく入れ物。**在庫は確保しない**。カートに入れただけでは誰の在庫でもない | 買い物かご（コード・API では不可）、確保、キープ |
| チェックアウトセッション | `CheckoutSession` / `checkout_sessions` | `POST /api/checkout-sessions`, `checkout_session_id` | 購入手続き | 「購入手続きに進む」から注文確定までの 1 回の手続き。在庫ソフトロックと決済インテントをこれが束ねる | チェックアウト（単独）、購入セッション、オーダードラフト |
| 販売注文 | `SellingOrder` / `selling_orders` | `/api/selling-orders`, `selling_order_id` | 注文 | 顧客が商品を買った取引 1 件。Stripe の Webhook を受けて初めて作られる | 購入注文、オーダー、受注、`order`（単独） |
| 注文明細 | `SellingOrderItem` / `selling_order_items` | `selling_order_items` | 注文内容 | 販売注文に含まれる個体 1 点ぶんの記録。どの在庫がいくらで売れたかを確定させる | `selling_order_inventories`、注文商品、カート明細 |
| 販売注文ステータス | `SellingOrderStatus` | `selling_order_status` | ステータス | 販売注文の進行状況。許可された遷移だけを通す（状態遷移表を参照） | 注文状態、進捗、`status`（単独） |
| 決済 | `Payment` / `payments` | `payment_id` | お支払い | 顧客から代金を受け取ること。買取の送金とは向きが逆で、別の概念として扱う | 支払い（買取と紛らわしいので不可）、入金、決済処理 |
| 決済インテント | `PaymentIntent` | `stripe_payment_intent_id` | （画面に出さない） | Stripe 側で 1 回の支払いを表すオブジェクトの ID。チェックアウトセッションと 1:1 | Stripe ID、`payment_id`（Stripe のものを指して使うのは不可） |
| 決済 Webhook イベント | `StripeWebhookEvent` / `stripe_webhook_events` | `POST /api/webhooks/stripe` | （画面に出さない） | Stripe から届く通知。**注文確定はこれを正とする**。同じイベントが複数回届く前提で、イベント ID で重複を弾く | コールバック、通知、Stripe 連携 |
| 注文金額の内訳 | `Subtotal` / `ShippingFee` / `Tax` / `TotalAmount` | `subtotal`, `shipping_fee`, `tax_amount`, `total_amount` | 小計 / 配送料 / 消費税 / 合計 | 販売注文の金額の構成。合計だけでなく内訳も保存し、後から計算し直さない | `product_all_price`、総額、税込金額 |
| 配送 | `Shipment` / `shipments` | `/api/admin/selling-orders/{id}/shipment` | 配送 | 販売注文の商品を顧客に届ける手配 1 件 | 発送（動作としてのみ可）、デリバリー、出荷 |
| 配送業者 | `Carrier` | `carrier_code` | 配送業者 | 荷物を運ぶ会社。ヤマト家財便・佐川急便・日本郵便など | `delivery_company_code`、運送会社、キャリア（外部連携の意味と紛らわしい） |
| 追跡番号 | `TrackingNumber` | `tracking_number` | 追跡番号 | 配送業者が発行する荷物の問い合わせ番号 | 伝票番号、お問い合わせ番号（UI 文言としてのみ可） |
| 配送先住所 | `ShippingAddress` / `shipping_addresses` | `shipping_address_id` | お届け先 | 販売注文ごとに確定して保存する届け先。顧客の登録住所を後から変えても、過去の注文の届け先は変わらない | 住所（単独）、配送先、`address`（単独） |
| 注文キャンセル | `Cancellation` | `POST /api/admin/selling-orders/{id}/cancellation`, `canceled_at` | キャンセル | 販売注文を取り消すこと。在庫は販売可能に戻さず、返品入庫として別途扱う | 取消、無効化、削除 |

### 3.5 買取

| 日本語の用語 | コード上の名前 | API 上の名前 | UI 文言 | 定義 | 使ってはいけない別名 |
| --- | --- | --- | --- | --- | --- |
| 買取注文 | `PurchaseOrder` / `purchase_orders` | `/api/purchase-orders`, `purchase_order_id` | 買取依頼 | 顧客が売却を申し込んでから送金が終わるまでの取引 1 件。MVP では 1 依頼 1 点とする | 買取依頼（コード・API では不可）、`BuyingOrder`、仕入注文、買取オーダー |
| 買取注文ステータス | `PurchaseOrderStatus` | `purchase_order_status` | ステータス | 買取注文の進行状況。許可された遷移だけを通す（状態遷移表を参照） | 査定ステータス、フェーズ、`status`（単独） |
| 申告 | `SelfDeclaration` | `declared_rank`, `declared_note` | お客様の申告 | 顧客が申込時に自分で書いた状態の説明と自己申告のランク。運営の査定結果とは必ず区別する | `seller_description`、自己査定、ユーザーメモ |
| 写真査定 | `PhotoAssessment` | `POST /api/admin/purchase-orders/{id}/photo-assessment` | 写真査定 | 顧客が送った写真だけを見て暫定の金額を出す工程。現物は見ていない | 仮査定、一次査定、簡易査定 |
| 暫定査定額 | `EstimatedAmount` | `estimated_amount` | 参考見積もり | 写真査定で出した金額。顧客に対する確定した約束ではない | 見積額、仮金額、`estimate` |
| 本人確認 | `IdentityVerification` | `/api/admin/purchase-orders/{id}/identity-verification` | 本人確認（eKYC） | 古物営業法のために、身分証と申込内容の氏名・住所が一致するか確認する工程 | eKYC（コード・API では不可）、KYC、身元確認 |
| 集荷 | `Pickup` | `pickup_scheduled_date`, `pickup_time_slot` | 集荷 | 配送業者が顧客のもとへ商品を取りに行くこと | `collect_method`、回収、引き取り |
| 集荷料 | `PickupFee` | `pickup_fee` | 集荷料金 | 集荷にかかった費用。最終査定額から差し引いてから送金する | 送料、手数料、`total_collect_price` |
| 実物検品 | `Inspection` | `POST /api/admin/purchase-orders/{id}/inspection` | 検品 | 届いた現物を運営が確認し、コンディションランクを確定させる工程 | 検査、査定（単独）、チェック |
| 減額 | `Deduction` / `purchase_order_deductions` | `deduction_reason`, `deduction_amount` | 減額 | 実物検品で見つかった不具合ごとに、暫定査定額から差し引く金額とその理由。写真 1 枚ごとに紐づく | 値引き、ペナルティ、`penalty`、割引 |
| 最終査定額 | `FinalAssessedAmount` | `final_assessed_amount` | 最終査定金額 | 実物検品の減額を反映して確定した買取金額。集荷料を差し引く前の額 | 買取価格、`final_purchase_price`、確定額、落札額 |
| 合意 | `Agreement` | `POST /api/purchase-orders/{id}/agreement`, `agreed_at` | 査定金額に合意する | 顧客が最終査定額を承諾すること。これが成立するまで送金も再出品もしない | 承認、同意（UI 文言としてのみ可）、確定 |
| 返送 | `Return` | `POST /api/purchase-orders/{id}/return` | 返品を依頼 | 顧客が最終査定額に合意せず、商品を返してもらうこと。返送料は請求しない | 返品（UI 文言としてのみ可）、キャンセル、差し戻し |
| 送金 | `Payout` / `payouts` | `POST /api/admin/purchase-orders/{id}/payout` | 振込 | 合意済みの買取金額を Stripe 経由で顧客の口座へ支払うこと。販売側の決済とは向きが逆 | 支払い、入金、振込（コード・API では不可）、返金 |
| 振込先口座 | `BankAccount` / `bank_accounts` | `bank_account_id` | 振込先口座 | 顧客が送金を受け取る銀行口座。金融機関・支店・種別・口座番号・名義を持つ | 口座情報、銀行情報、`account`（単独） |

### 3.6 外部連携

| 日本語の用語 | コード上の名前 | API 上の名前 | UI 文言 | 定義 | 使ってはいけない別名 |
| --- | --- | --- | --- | --- | --- |
| Stripe 決済連携 | `StripePaymentGateway` | `POST /api/webhooks/stripe` | （画面に出さない） | 販売の決済を Stripe に任せる連携。MVP のスコープ内 | 決済代行、ペイメント、PSP |
| Stripe 送金連携 | `StripePayoutGateway` | — | （画面に出さない） | 買取金を顧客口座へ送る連携。決済連携とは別の窓口として扱う | 振込 API、送金代行 |
| freee 連携 | `FreeeClient` | — | 会計連携 | 仕訳データを会計ソフトへ送る連携。**MVP 対象外**。用語だけ予約する | 会計、経理連携、`accounting_integration` |
| Zendesk 連携 | `ZendeskClient` | — | お問い合わせ | 問い合わせチケットを扱う連携。**MVP 対象外**。用語だけ予約する | サポート、CS 連携、チケット |
| 配送キャリア API 連携 | `CarrierClient` | — | 送り状発行 | 送り状と追跡番号を配送業者の API から取る連携。**MVP 対象外**。MVP は管理者が追跡番号を手入力する | 送り状 API、ヤマト連携、`shipping_api` |

### 3.7 MVP 対象外の予約語

実装しないが、後から別の名前で作り直さないために名前だけ決めておく。

| 日本語の用語 | コード上の名前 | 定義 | 使ってはいけない別名 |
| --- | --- | --- | --- |
| 減価償却シミュレーター | `DepreciationSimulator` | 購入額と製造年から、経費算入額と減価償却を試算する機能 | 節税シミュレーター、税務計算 |
| 隠蔽家具提案 | `ConcealableFurnitureSuggestion` | 仕事用機材を隠せる家具を提案する機能 | 週末モード、収納提案 |
| 互換性チェッカー | `CompatibilityChecker` | MacBook Pro のモデルと周辺機器の仕様を照合する機能 | スペック判定、対応表 |
| ポイント | `Point` | 購入に応じて顧客へ付与する残高 | マイル、リワード |
| クーポン | `Coupon` | 次回購入時に使える割引券 | 割引コード、プロモコード |
| 買取明細書 | `PurchaseStatement` | 確定申告に使える買取内容の証明書類 | 領収書、資産証明 |

## 4. 使ってはいけない別名 早見表

書いてしまいがちな名前と、正しい名前の対応。

| 使ってはいけない名前 | 正しい名前 | 備考 |
| --- | --- | --- |
| ユーザー / `users` | 顧客 / `customers` | 管理者と紛らわしいため |
| 商品（コード・API 上） | 販売枠 `SellingProduct` / 在庫 `Inventory` | 「売り出す単位」と「実在する 1 点」は別物。UI 文言としての「商品」は可 |
| 在庫数 | 在庫 `Inventory` | 1 点ものなので数量という概念を持たない |
| `is_rock` / `is_temporary` | 在庫ソフトロック `InventoryLock` | 真偽値では「誰がいつまで」を表現できず、二重販売を防げない |
| 買取依頼（コード・API 上） | 買取注文 `PurchaseOrder` | UI 文言としての「買取依頼」は可 |
| `BuyingOrder` | `PurchaseOrder` | 既存の管理画面資料との不一致をこちらに寄せた |
| 支払い | 決済 `Payment` / 送金 `Payout` | 販売と買取でお金の向きが逆なので必ず区別する |
| 振込（コード・API 上） | 送金 `Payout` | UI 文言としての「振込」は可 |
| eKYC（コード・API 上） | 本人確認 `IdentityVerification` | UI 文言としての「eKYC」は可 |
| 査定（単独） | 写真査定 `PhotoAssessment` / 実物検品 `Inspection` | どちらの工程か必ず明示する |
| `final_purchase_price` | 最終査定額 `final_assessed_amount` | 送金額（集荷料を引いた後）と混同しないため |
| グレード / ランク（単独） | コンディションランク `ConditionRank` | 自己申告ランクとも区別する |
| `saved_place` | 保管場所 `storage_location` | — |
| `collect_*` | 集荷 `pickup_*` | — |
| `delivery_company_code` | 配送業者 `carrier_code` | — |
| `product_images_1` 〜 `_10` | 商品画像 `selling_product_images` | 枚数可変の別テーブルに正規化した |
| `selling_order_inventories` | 注文明細 `selling_order_items` | — |
| `product_all_price` | 合計 `total_amount` | 内訳（小計・配送料・消費税）も別カラムで保存する |
