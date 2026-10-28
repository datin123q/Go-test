-- +goose Up
alter table blog_post
    add published_at datetime null after write_at;

-- +goose Down
alter table blog_post
    drop column published_at;

