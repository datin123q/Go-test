-- +goose Up
alter table blog_post
    modify id varchar(50) not null;

-- +goose Down
alter table energysharesdb.blog_post
    modify id BIGINT auto_increment;

alter table energysharesdb.blog_post
    auto_increment = 100;
