-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS links (
    original_url TEXT NOT NULL,
    short_code VARCHAR(10) NOT NULL,

    CONSTRAINT links_short_code_unique PRIMARY KEY (short_code),
    CONSTRAINT links_original_url_unique UNIQUE (original_url)
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS links;
-- +goose StatementEnd