-- +goose Up
alter table project
    drop
        column site_characteristics;

alter table project
    drop
        column climatology_and_hazards;

alter table project
    drop
        column technical_aspects;

alter table project
    drop
        column equipment_recognition_and_system_access_data;

alter table project
    drop
        column local_permits;

alter table project
    drop
        column land_status;

alter table project
    drop
        column total_evaluation_score;

alter table project
    drop
        column about_evaluation;

-- +goose Down
alter table project
    add site_characteristics float null;

alter table project
    add climatology_and_hazards float null;

alter table project
    add technical_aspects float null;

alter table project
    add equipment_recognition_and_system_access_data float null;

alter table project
    add local_permits float null;

alter table project
    add land_status float null;

alter table project
    add total_evaluation_score varchar(255) null;

alter table project
    add about_evaluation longtext null;