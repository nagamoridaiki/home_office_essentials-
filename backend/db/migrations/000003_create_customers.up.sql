BEGIN;

CREATE TABLE IF NOT EXISTS customers (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- 小文字にそろえた値だけを入れる。UNIQUE が大文字・小文字違いの重複も弾く
    email text NOT NULL UNIQUE CHECK (email = lower(email)),
    -- 氏名・フリガナ・電話番号はアカウント作成時は NULL。プロフィール編集で登録する
    family_name text,
    given_name text,
    family_name_kana text,
    given_name_kana text,
    phone_number text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE OR REPLACE TRIGGER customers_set_updated_at
    BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
