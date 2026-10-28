-- +goose Up
alter table project
drop column banner_image_url;

alter table project
drop column banner_image_id;

alter table project
drop column public_view_end_at;

alter table project
drop column is_achieve_minimum_fundraising;



-- +goose Down
alter table project
    add public_view_end_at datetime null;

alter table project
    add is_achieve_minimum_fundraising tinyint 0;