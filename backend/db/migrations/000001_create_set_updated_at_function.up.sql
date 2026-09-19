BEGIN;

-- 行が更新されるたびに updated_at を現在時刻にする。各テーブルの BEFORE UPDATE トリガーから呼ぶ
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMIT;
