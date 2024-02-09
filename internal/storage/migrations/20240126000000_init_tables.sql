-- +goose Up
-- +goose StatementBegin
CREATE table whitelist (
    ip              numeric,
    mask            numeric,
    ipfrom          numeric,
    ipto            numeric);
CREATE table blacklist (
    ip              numeric,
    mask            numeric,
    ipfrom          numeric,
    ipto            numeric);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table whitelist;
drop table blacklist;
-- +goose StatementEnd

