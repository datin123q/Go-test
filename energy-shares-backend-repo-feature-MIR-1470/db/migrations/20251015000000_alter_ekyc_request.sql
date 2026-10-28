-- +goose Up
alter table ekyc_request_log add result_response json;

-- +goose Down
alter table ekyc_request_log drop column result_response;
