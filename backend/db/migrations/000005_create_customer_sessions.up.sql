BEGIN;

CREATE TABLE IF NOT EXISTS customer_sessions (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id bigint NOT NULL REFERENCES customers (id) ON DELETE RESTRICT,
    -- Cookie に入れたトークンの SHA-256。生の値は保存しない
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    -- ログアウトした時刻。行は消さない
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS customer_sessions_customer_id_idx
    ON customer_sessions (customer_id);

CREATE OR REPLACE TRIGGER customer_sessions_set_updated_at
    BEFORE UPDATE ON customer_sessions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

COMMIT;
