-- +goose Up
-- +goose StatementBegin
CREATE table whitelist (
    ip              cidr);
CREATE table blacklist (
    ip              cidr);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table whitelist;
drop table blacklist;
-- +goose StatementEnd

