-- +goose Up
alter table holiday
    add title varchar(200) null after id;

alter table holiday
    drop column year;

alter table holiday
    drop column month;

alter table holiday
    drop column day;

alter table holiday
    modify date varchar(50) null;

alter table holiday
    modify holiday_date datetime null after date;


-- +goose Down
-- +goose StatementBegin
alter table holiday

drop column title;

alter table holiday
    modify date varchar(255) null;

alter table holiday
    add year int null;

alter table holiday
    add day int null;

alter table holiday
    add month int null;
-- +goose StatementEnd
