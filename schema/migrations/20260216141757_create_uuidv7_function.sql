-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION uuidv7() RETURNS uuid AS $$
DECLARE
    v_ts  bigint  := (EXTRACT(epoch FROM clock_timestamp()) * 1000)::bigint;
    v_rnd bytea   := substring(uuid_send(gen_random_uuid()) || uuid_send(gen_random_uuid()), 1, 10);
    v_b0  integer := (v_ts >> 40) & 255;
    v_b1  integer := (v_ts >> 32) & 255;
    v_b2  integer := (v_ts >> 24) & 255;
    v_b3  integer := (v_ts >> 16) & 255;
    v_b4  integer := (v_ts >>  8) & 255;
    v_b5  integer :=  v_ts        & 255;
    v_b6  integer := 112 | (get_byte(v_rnd, 0) & 15);  -- version nibble = 7
    v_b7  integer := get_byte(v_rnd, 1);
    v_b8  integer := 128 | (get_byte(v_rnd, 2) & 63);  -- variant bits = 10
    v_b9  integer := get_byte(v_rnd, 3);
    v_b10 integer := get_byte(v_rnd, 4);
    v_b11 integer := get_byte(v_rnd, 5);
    v_b12 integer := get_byte(v_rnd, 6);
    v_b13 integer := get_byte(v_rnd, 7);
    v_b14 integer := get_byte(v_rnd, 8);
    v_b15 integer := get_byte(v_rnd, 9);
BEGIN
    RETURN (
        lpad(to_hex(v_b0),  2, '0') ||
        lpad(to_hex(v_b1),  2, '0') ||
        lpad(to_hex(v_b2),  2, '0') ||
        lpad(to_hex(v_b3),  2, '0') || '-' ||
        lpad(to_hex(v_b4),  2, '0') ||
        lpad(to_hex(v_b5),  2, '0') || '-' ||
        lpad(to_hex(v_b6),  2, '0') ||
        lpad(to_hex(v_b7),  2, '0') || '-' ||
        lpad(to_hex(v_b8),  2, '0') ||
        lpad(to_hex(v_b9),  2, '0') || '-' ||
        lpad(to_hex(v_b10), 2, '0') ||
        lpad(to_hex(v_b11), 2, '0') ||
        lpad(to_hex(v_b12), 2, '0') ||
        lpad(to_hex(v_b13), 2, '0') ||
        lpad(to_hex(v_b14), 2, '0') ||
        lpad(to_hex(v_b15), 2, '0')
    )::uuid;
END;
$$ LANGUAGE plpgsql VOLATILE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS uuidv7();
-- +goose StatementEnd
