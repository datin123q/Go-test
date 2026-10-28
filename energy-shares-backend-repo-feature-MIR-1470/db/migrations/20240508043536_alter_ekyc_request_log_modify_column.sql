-- +goose Up
-- ekyc_request_log 테이블에 user_id 컬럼 추가
alter table ekyc_request_log
    add user_id bigint not null after email;

-- ekyc_request_log 테이블의 user_id 컬럼에 값을 채워넣기
update ekyc_request_log
set user_id = ifnull((select private_profile.user_id
                      from private_profile
                      where private_profile.request_id = ekyc_request_log.request_id), 0)
where user_id = 0;

-- ekyc_request_log 테이블의 request_id 컬럼에 있던 인덱스 제거
drop index ekyc_request_log_request_id_index on ekyc_request_log;

-- ekyc_request_log 테이블의 user_id 컬럼에 인덱스 추가
create index idx_ekyc_request_log_user_id
    on ekyc_request_log (user_id);


-- ekyc_request_log 테이블의 request_id 컬럼 제거
alter table ekyc_request_log
drop
column request_id;

-- +goose Down
drop index idx_ekyc_request_log_user_id on ekyc_request_log;

alter table ekyc_request_log
drop
column user_id;

alter table ekyc_request_log
add request_id varchar(255) not null after email;

create index idx_ekyc_request_log_request_id
    on ekyc_request_log (request_id);