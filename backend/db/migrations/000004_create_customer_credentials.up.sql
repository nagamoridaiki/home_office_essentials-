BEGIN;

CREATE TABLE IF NOT EXISTS customer_credentials (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- 認証情報は顧客 1 人に 1 件
    customer_id bigint NOT NULL UNIQUE REFERENCES customers (id) ON DELETE RESTRICT,
    -- 自前認証用。bcrypt のハッシュを入れ、平文は保存しない
    password_hash text,
    -- 外部の認証サービス（OIDC）の iss と sub。自前認証のあいだは NULL
    auth_issuer text,
    auth_subject text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (auth_issuer, auth_subject)
);

CREATE OR REPLACE TRIGGER customer_credentials_set_updated_at
    BEFORE UPDATE ON customer_credentials
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
