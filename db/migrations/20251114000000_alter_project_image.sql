-- +goose Up
alter table project_image add `order` int not null default 0 AFTER image_type;

-- +goose Down
alter table project_image drop column `order`;
