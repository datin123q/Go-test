create table if not exists admin_profile (
    id bigint auto_increment primary key,
    user_id bigint null,
    validity_start_date datetime null,
    validity_end_date datetime null,
    role_updated datetime null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists answer (
    id bigint auto_increment primary key,
    qna_id bigint not null,
    common_id varchar(40) not null,
    answer varchar(255) null,
    created datetime null,
    updated datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists antisocial_check_file (
    id bigint auto_increment primary key,
    check_id bigint not null,
    file_name varchar(255) not null,
    original_name varchar(255) not null,
    file_size bigint not null,
    file_extension varchar(255) not null,
    file_path varchar(255) not null,
    file_url varchar(255) not null,
    owner_id bigint not null,
    update_id bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_antisocial_check_file_file_name unique (file_name)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists attachment (
    id bigint auto_increment primary key,
    file_name varchar(255) not null,
    original_name varchar(255) not null,
    file_size bigint not null,
    file_extension varchar(255) not null,
    file_path varchar(255) not null,
    file_url varchar(255) not null,
    resource_id varchar(40) not null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_attachment_file_name unique (file_name)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_attachment_resource_id on attachment (resource_id);
create table if not exists audit_status_history (
    id bigint auto_increment primary key,
    user_id bigint null,
    owner_id bigint null,
    update_id bigint null,
    processor varchar(255) not null,
    state varchar(255) not null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_audit_status_history_owner_id on audit_status_history (owner_id);
create index IDX_audit_status_history_update_id on audit_status_history (update_id);
create table if not exists blog_post (
    id bigint auto_increment primary key,
    title varchar(255) not null,
    link varchar(255) not null,
    data text null,
    category varchar(255) null,
    source_url varchar(255) null,
    write_at datetime null,
    created datetime null,
    updated datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists casbin_rule (
    p_type varchar(100) null,
    v0 varchar(100) null,
    v1 varchar(100) null,
    v2 varchar(100) null,
    v3 varchar(100) null,
    v4 varchar(100) null,
    v5 varchar(100) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_casbin_rule_p_type on casbin_rule (p_type);
create index IDX_casbin_rule_v0 on casbin_rule (v0);
create index IDX_casbin_rule_v1 on casbin_rule (v1);
create index IDX_casbin_rule_v2 on casbin_rule (v2);
create index IDX_casbin_rule_v3 on casbin_rule (v3);
create index IDX_casbin_rule_v4 on casbin_rule (v4);
create index IDX_casbin_rule_v5 on casbin_rule (v5);
create table if not exists `column` (
    id bigint auto_increment primary key,
    title varchar(255) not null,
    content longtext not null,
    user_id bigint not null,
    thumbnail_url varchar(255) not null,
    created datetime not null,
    updated datetime null,
    deleted datetime null,
    common_id varchar(40) not null,
    is_posted tinyint default 0 not null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists column_file (
    id bigint auto_increment primary key,
    column_id bigint not null,
    user_id bigint not null,
    file_name varchar(255) not null,
    original_name varchar(255) not null,
    file_type varchar(20) not null,
    file_size bigint not null,
    file_extension varchar(20) not null,
    file_path varchar(255) not null,
    file_url varchar(255) not null,
    created datetime not null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index column_file_column_id_index on column_file (column_id);
create table if not exists consult_note (
    id bigint auto_increment primary key,
    user_id bigint not null,
    owner_id bigint not null,
    modifier_id bigint not null,
    owner varchar(50) not null,
    title varchar(255) not null,
    content longtext not null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_consult_note_user_id on consult_note (user_id);
create table if not exists contact (
    id bigint auto_increment primary key,
    name varchar(255) null,
    email varchar(255) null,
    title varchar(255) null,
    contents longtext null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists content_image (
    id bigint auto_increment primary key,
    file_name varchar(255) not null,
    original_name varchar(255) not null,
    original_url varchar(255) null,
    file_size bigint not null,
    file_extension varchar(255) not null,
    file_path varchar(255) not null,
    file_url varchar(255) not null,
    resource_id varchar(40) not null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_content_image_file_name unique (file_name)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_content_image_resource_id on content_image (resource_id);
create table if not exists contract (
    id bigint auto_increment primary key,
    share_id varchar(255) null,
    created datetime null,
    deleted datetime null,
    share_qty bigint null,
    share_amount bigint null,
    status varchar(255) null,
    project_id bigint null,
    user_name varchar(255) null,
    contract_after_page_url varchar(255) null,
    updated datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_contract_share_id on contract (share_id);
create table if not exists corporate_profile (
    id bigint auto_increment primary key,
    user_id bigint null,
    corp_no varchar(255) not null,
    corp_type varchar(255) not null,
    employee_count varchar(255) not null,
    capital_amount varchar(255) not null,
    sales varchar(255) not null,
    representative_family_name varchar(255) not null,
    representative_first_name varchar(255) not null,
    representative_family_name_kana varchar(255) not null,
    representative_first_name_kana varchar(255) not null,
    representative_mobile_phone varchar(255) not null,
    representative_email varchar(255) not null,
    manager_family_name varchar(255) not null,
    manager_first_name varchar(255) not null,
    manager_family_name_kana varchar(255) not null,
    manager_first_name_kana varchar(255) not null,
    manager_email varchar(255) not null,
    manager_zip_code varchar(255) not null,
    manager_province varchar(255) not null,
    manager_city varchar(255) not null,
    manager_address varchar(255) not null,
    manager_address2 varchar(255) null,
    manager_mobile_phone varchar(255) not null,
    manager_birthday datetime not null,
    is_manager_sms_cert tinyint(1) null,
    manager_sms_cert_at datetime null,
    department varchar(255) not null,
    position varchar(255) not null,
    corporate_postal_code varchar(255) null,
    corporate_postal_code_approval_date datetime null,
    corporate_postal_status tinyint(1) null,
    corporate_postal_code_send_date datetime null,
    is_representative_sms_cert tinyint(1) null,
    representative_sms_cert_at datetime null,
    manager_is_address_change tinyint(1) null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_corporate_profile_corp_no unique (corp_no),
    constraint UQE_corporate_profile_user_id unique (user_id)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists corporate_shareholder (
    id bigint auto_increment primary key,
    user_id bigint not null,
    uid varchar(255) null,
    family_name varchar(255) not null,
    first_name varchar(255) not null,
    family_name_kana varchar(255) not null,
    first_name_kana varchar(255) not null,
    birthday datetime not null,
    zip_code varchar(255) null,
    address varchar(255) null,
    address2 varchar(255) null,
    province varchar(255) null,
    city varchar(255) not null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists corporate_sign_up_request (
    id bigint auto_increment primary key,
    request_id varchar(40) not null,
    email varchar(255) null,
    url varchar(255) null,
    expire datetime null,
    created datetime null,
    send_count int null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists deposit_mapping (
    id bigint auto_increment primary key,
    request_id bigint null,
    va_deposit_id bigint null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists deposit_request (
    id bigint auto_increment primary key,
    user_id bigint null,
    share_id varchar(255) not null,
    project_id bigint not null,
    project_type varchar(255) null,
    request_amount bigint null,
    deposit_amount bigint null,
    refund_amount bigint null,
    deposit_date datetime null,
    deposit_check_status varchar(255) null,
    update_date datetime null,
    expire_deposit_date datetime null,
    deposit_bank_name varchar(255) null,
    deposit_branch_name varchar(255) null,
    deposit_branch_code varchar(255) null,
    deposit_account_number varchar(255) null,
    created datetime null,
    deleted datetime null,
    constraint UQE_deposit_request_share_id unique (share_id)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_deposit_request_user_id on deposit_request (user_id);
create table if not exists deposit_request_lack_history (
    id bigint auto_increment primary key,
    user_id bigint null,
    deposit_request_id bigint null,
    request_execute_date datetime null,
    deposit_expected_date datetime null,
    amount bigint null,
    account_number varchar(255) null,
    project_type varchar(255) null,
    project_id bigint null,
    share_id varchar(255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists deposit_withdrawal_execute_history (
    id bigint auto_increment primary key,
    investor_id bigint null,
    investor_name varchar(255) null,
    account_number varchar(255) null,
    amount bigint null,
    operator_name varchar(255) null,
    operator_email varchar(255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists display_code (
    id bigint auto_increment primary key,
    code varchar(255) null,
    active tinyint(1) null,
    created datetime null,
    updated datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists dividend (
    id bigint auto_increment primary key,
    user_id bigint null,
    project_id bigint null,
    project_type varchar(255) null,
    settlement_id bigint null,
    investment_amount bigint null,
    investment_qty bigint null,
    amount bigint null,
    qty bigint null,
    deposit_amount bigint null,
    tax bigint null,
    dividend_date datetime null,
    transaction_status varchar(255) null,
    transaction_type varchar(255) null,
    transaction_date datetime null,
    transaction_id varchar(40) not null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    confirmed_date datetime null,
    spv_id bigint null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_dividend_transaction_id on dividend (transaction_id);
create table if not exists ekyc_request_log (
    id bigint auto_increment primary key,
    email varchar(255) not null,
    request_id varchar(40) not null,
    ekyc_id varchar(40) not null,
    applicant_id bigint not null,
    status varchar(10) not null,
    created datetime not null,
    result varchar(10) null,
    error_message longtext null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index ekyc_request_log_request_id_index on ekyc_request_log (request_id);
create table if not exists email_code_request (
    id bigint auto_increment primary key,
    email varchar(255) null,
    code varchar(255) null,
    expire datetime null,
    created datetime null,
    send_count int null,
    user_id bigint null,
    first_req tinyint(1) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists event_message (
    event_category varchar(255) null,
    content varchar(500) null,
    uuid varchar(40) not null primary key,
    event_time datetime null,
    is_publish tinyint(1) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists faq (
    id bigint auto_increment primary key,
    common_id varchar(40) not null,
    pin tinyint(1) not null,
    title varchar(255) not null,
    owner_id bigint null,
    update_id bigint null,
    content_html longtext not null,
    posted tinyint(1) null,
    closing_date datetime not null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    attachment_cnt int null,
    category_type varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_faq_owner_id on faq (owner_id);
create index IDX_faq_update_id on faq (update_id);
create table if not exists funding_request (
    id bigint auto_increment primary key,
    common_id varchar(40) not null,
    name varchar(255) not null,
    zip_code varchar(255) not null,
    address varchar(255) not null,
    address2 varchar(255) null,
    corp_no varchar(255) not null,
    province varchar(255) not null,
    city varchar(255) not null,
    email varchar(255) not null,
    phone varchar(30) not null,
    equipment_name varchar(255) null,
    equipment_address varchar(255) null,
    equipment_address2 varchar(255) null,
    request_status varchar(255) null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    equipment_info varchar(400) null,
    equipment_consulting varchar(400) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists funding_request_history (
    id bigint auto_increment primary key,
    funding_request_id bigint null,
    substance_type varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    content_type varchar(255) null,
    prev_content varchar(255) null,
    curr_content varchar(255) null,
    prev_url varchar(255) null,
    curr_url varchar(255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists gmo_aozora_client (
    id bigint auto_increment primary key,
    client_id varchar(255) null,
    client_secret varchar(255) null,
    scope varchar(255) null,
    redirect_uri varchar(255) null,
    access_token varchar(255) null,
    refresh_token varchar(255) null,
    created datetime null,
    updated datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists holiday (
    id bigint auto_increment primary key,
    year int null,
    month int null,
    day int null,
    date varchar(255) null,
    created datetime null,
    holiday_date datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists investment_request (
    id bigint auto_increment primary key,
    investment_waiters_id bigint null,
    request_id varchar(40) not null,
    email varchar(255) null,
    url varchar(255) null,
    qty bigint null,
    is_expire tinyint(1) null,
    created datetime null,
    project_id bigint null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists investment_wait_history (
    id bigint auto_increment primary key,
    investment_wait_id bigint null,
    substance_type varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    prev_status varchar(255) null,
    curr_status varchar(255) null,
    mail_send_date datetime null,
    process_method varchar(255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists investment_waiters (
    id bigint auto_increment primary key,
    project_id bigint null,
    user_id bigint null,
    waiting_turn bigint null,
    qty bigint null,
    amount bigint null,
    status varchar(255) null,
    waiting_date datetime null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    user_investor_type varchar(255) null,
    user_email varchar(255) null,
    user_name varchar(255) null,
    user_mobile_phone varchar(255) null,
    share_id varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists job_execute_time (
    id bigint auto_increment primary key,
    deposit_check_time datetime null,
    va_account_issue_time datetime null,
    transfer_request_time datetime null,
    transfer_result_time datetime null,
    deposit_process_time datetime null,
    transfer_status_time datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists login_history (
    id bigint auto_increment primary key,
    user_id bigint null,
    name varchar(255) not null,
    email varchar(255) not null,
    ip_addr varchar(255) null,
    agent varchar(255) null,
    login_result varchar(255) null,
    login_time datetime null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_login_history_user_id on login_history (user_id);
create table if not exists mail (
    id bigint auto_increment primary key,
    status varchar(255) null,
    template_type varchar(255) null,
    subject varchar(255) null,
    name varchar(255) null,
    sender varchar(255) null,
    content longtext null,
    mail_type varchar(255) null,
    count bigint null,
    send_date datetime null,
    updated datetime null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists mail_request (
    id bigint auto_increment primary key,
    mail_id bigint null,
    status varchar(255) null,
    receiver varchar(255) null,
    template_data longtext null,
    message_id varchar(255) null,
    message_status varchar(255) null,
    message_error varchar(255) null,
    send_date datetime null,
    updated datetime null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_mail_request_mail_id on mail_request (mail_id);
create table if not exists map_mail_receiver (
    id bigint auto_increment primary key,
    mail_id int not null,
    receiver varchar(50) not null,
    created datetime not null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists map_sms_receiver (
    id bigint auto_increment primary key,
    sms_id int not null,
    receiver varchar(50) not null,
    created datetime not null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists marketing_code (
    id bigint auto_increment primary key,
    code varchar(255) not null,
    campaign varchar(255) not null,
    sign_up_request_id varchar(255) not null,
    email varchar(255) not null,
    created datetime null,
    updated datetime null,
    user_id bigint null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists mobile_phone_verification_request (
    id bigint auto_increment primary key,
    request_id varchar(40) not null,
    user_id bigint null,
    code varchar(255) null,
    mobile_phone varchar(255) null,
    status varchar(255) null,
    expire datetime null,
    created datetime null,
    send_count int null,
    first_req tinyint(1) null,
    phone_type varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists my_number (
    id bigint auto_increment primary key,
    common_id varchar(40) not null,
    user_id bigint not null,
    status varchar(255) not null,
    my_number varchar(255) not null,
    manager_name varchar(255) null,
    approval_date datetime null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists news (
    id bigint auto_increment primary key,
    common_id varchar(40) not null,
    pin tinyint(1) not null,
    title varchar(255) not null,
    owner_id bigint null,
    update_id bigint null,
    content_html longtext not null,
    posted tinyint(1) null,
    closing_date datetime not null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    attachment_cnt int null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_news_owner_id on news (owner_id);
create index IDX_news_update_id on news (update_id);
create table if not exists notice_msg (
    id bigint auto_increment primary key,
    title varchar(255) null,
    content_html longtext null,
    notice_date datetime null,
    notice_user_count bigint null,
    created datetime null,
    notice_user_type varchar(255) null,
    project_id bigint null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists operator_history (
    id bigint auto_increment primary key,
    user_id bigint not null,
    name varchar(255) null,
    email varchar(255) null,
    menu varchar(255) not null,
    ip varchar(255) not null,
    title varchar(255) null,
    state varchar(255) not null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists password_reset_request (
    id bigint auto_increment primary key,
    request_id varchar(40) not null,
    email varchar(255) null,
    url varchar(255) null,
    expire datetime null,
    created datetime null,
    send_count int null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists policy (
    id bigint auto_increment primary key,
    common_id varchar(40) not null,
    owner_id bigint null,
    update_id bigint null,
    policy_type varchar(255) not null,
    policy_view_type varchar(255) not null,
    content_html longtext not null,
    updator_email varchar(255) not null,
    attachment_id bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_policy_owner_id on policy (owner_id);
create index IDX_policy_update_id on policy (update_id);
create table if not exists policy_info_history (
    id bigint auto_increment primary key,
    policy_id bigint null,
    substance_type varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    prev_content longtext null,
    curr_content longtext null,
    prev_content_html longtext null,
    curr_content_html longtext null,
    prev_url varchar(255) null,
    curr_url varchar(255) null,
    created datetime null,
    content_type varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists popular_info (
    id bigint auto_increment primary key,
    `index` int null,
    project_id bigint null,
    posted tinyint(1) null,
    project_name varchar(255) null,
    updated datetime null,
    created datetime null,
    constraint UQE_popular_info_index unique (`index`)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists private_profile (
    id bigint auto_increment primary key,
    user_id bigint null,
    birthday datetime not null,
    first_name varchar(255) not null,
    family_name varchar(255) not null,
    first_name_kana varchar(255) not null,
    family_name_kana varchar(255) not null,
    asset varchar(255) not null,
    annual_income varchar(255) not null,
    is_pro_invest_asset tinyint(1) not null,
    is_pro_invest_finance_asset tinyint(1) not null,
    nationality varchar(255) not null,
    job varchar(255) not null,
    gender varchar(255) not null,
    request_id varchar(255) not null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_private_profile_user_id unique (user_id)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists project (
    id bigint auto_increment primary key,
    common_id varchar(40) not null,
    spv_id bigint null,
    name varchar(100) not null,
    category_type varchar(255) not null,
    admin_note mediumtext null,
    project_info_html longtext null,
    company_info_html longtext null,
    description varchar(255) not null,
    start_at datetime null,
    end_at datetime null,
    project_type varchar(255) null,
    project_status varchar(255) null,
    owner_id bigint not null,
    update_id bigint null,
    capacity varchar(255) null,
    fund_internal_rate_return varchar(255) null,
    loan_interest_rate varchar(255) null,
    loan_term varchar(255) null,
    current_share_amount bigint null,
    current_share_qty bigint null,
    min_share_qty bigint null,
    share_price bigint null,
    target_min_share_amount bigint null,
    target_share_qty bigint null,
    target_share_amount bigint null,
    posted tinyint(1) null,
    block_invest tinyint(1) null,
    main_image_id bigint null,
    main_image_url varchar(255) null,
    banner_image_url varchar(255) null,
    banner_image_id bigint null,
    is_wait_invest tinyint(1) null,
    is_investor_limit tinyint(1) null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    site_characteristics float null,
    climatology_and_hazards float null,
    technical_aspects float null,
    equipment_recognition_and_system_access_data float null,
    local_permits float null,
    land_status float null,
    caution_info_html longtext null,
    bank_name varchar(255) null,
    branch_name varchar(255) null,
    bank_code varchar(255) null,
    branch_code varchar(255) null,
    bank_account varchar(255) null,
    depositor varchar(255) null,
    contract_before_page_date datetime null,
    contract_after_page_date datetime null,
    silent_partnerships_date datetime null,
    project_report_date datetime null,
    public_view_end_at datetime null,
    public_view_share_amount bigint null,
    is_manual_invest tinyint default 0 null,
    is_achieve_minimum_fundraising tinyint(1) default 0 not null,
    is_amount_limit tinyint(1) default 0 null,
    total_evaluation_score varchar(255) null,
    about_evaluation longtext null,
    project_progress_percent int null,
    is_visible_investment_info tinyint default 0 null,
    offering_type varchar(50) default 'OFFERING_PRIVATE' null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_project_spv_id on project (spv_id);
create table if not exists project_banner (
    id bigint auto_increment primary key,
    project_id bigint null,
    posted tinyint(1) null,
    `index` int null,
    project_name varchar(255) null,
    description varchar(255) null,
    banner_image_url varchar(255) null,
    post_date datetime null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists project_catalog (
    id bigint auto_increment primary key,
    project_id bigint not null,
    file_name varchar(255) not null,
    original_name varchar(255) not null,
    file_type varchar(255) null,
    file_size bigint not null,
    file_extension varchar(255) not null,
    file_path varchar(255) not null,
    file_url varchar(255) not null,
    owner_id bigint not null,
    update_id bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_project_catalog_file_name unique (file_name)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists project_changed_content (
    id bigint auto_increment primary key,
    info_id bigint null,
    content_type varchar(255) null,
    prev_content varchar(255) null,
    curr_content varchar(255) null,
    prev_content_html longtext null,
    curr_content_html longtext null,
    prev_url varchar(255) null,
    curr_url varchar(255) null,
    content_suffix varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists project_file (
    id bigint auto_increment primary key,
    project_id bigint not null,
    file_name varchar(255) not null,
    original_name varchar(255) not null,
    file_type varchar(255) null,
    file_size bigint not null,
    file_extension varchar(255) not null,
    file_path varchar(255) not null,
    file_url varchar(255) not null,
    owner_id bigint not null,
    update_id bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_project_file_file_name unique (file_name)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists project_image (
    id bigint auto_increment primary key,
    project_id bigint not null,
    file_name varchar(255) not null,
    original_name varchar(255) not null,
    file_type varchar(255) null,
    file_size bigint not null,
    file_extension varchar(255) not null,
    file_path varchar(255) not null,
    file_url varchar(255) not null,
    owner_id bigint not null,
    update_id bigint null,
    image_type varchar(255) not null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_project_image_file_name unique (file_name)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists project_info_history (
    id bigint auto_increment primary key,
    project_id bigint null,
    substance_type varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists project_investment_execute (
    id bigint auto_increment primary key,
    project_id bigint null,
    execute_status varchar(255) null,
    amount bigint null,
    execute_date datetime null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists project_milestone (
    id bigint auto_increment primary key,
    project_id bigint not null,
    milestone_status varchar(255) not null,
    milestone_no int not null,
    milestone_info varchar(255) not null,
    milestone_assignment int not null,
    milestone_note varchar(255) not null,
    created datetime null,
    updated datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_project_milestone_milestone_no on project_milestone (milestone_no);
create index IDX_project_milestone_milestone_status on project_milestone (milestone_status);
create index IDX_project_milestone_project_id on project_milestone (project_id);
create table if not exists project_news (
    id bigint auto_increment primary key,
    project_id bigint null,
    title varchar(255) null,
    owner_id bigint null,
    update_id bigint null,
    content_html longtext null,
    posted tinyint(1) null,
    post_date datetime null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_project_news_owner_id on project_news (owner_id);
create index IDX_project_news_update_id on project_news (update_id);
create table if not exists project_request (
    id bigint auto_increment primary key,
    user_id bigint null,
    user_email varchar(255) null,
    user_name varchar(255) null,
    user_type varchar(255) null,
    zip_code varchar(255) null,
    address varchar(255) null,
    address2 varchar(255) null,
    investor_type varchar(255) null,
    project_id bigint null,
    project_name varchar(255) null,
    request_date datetime null,
    response_date datetime null,
    response_status tinyint(1) null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_project_request_user_id on project_request (user_id);
create table if not exists project_view_count (
    id bigint auto_increment primary key,
    project_id bigint null,
    view_count bigint null,
    updated datetime null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_project_view_count_project_id on project_view_count (project_id);
create table if not exists qna (
    id bigint auto_increment primary key,
    common_id varchar(40) not null,
    user_id bigint not null,
    title varchar(255) not null,
    name varchar(255) not null,
    mobile_phone varchar(255) null,
    email varchar(255) not null,
    content varchar(255) not null,
    answer_count bigint default 0 not null,
    `read` tinyint(1) default 0 not null,
    created datetime null,
    updated datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists qna_info_history (
    id bigint auto_increment primary key,
    user_id bigint null,
    process_id bigint null,
    qna_id bigint null,
    content_type varchar(255) null,
    prev_content varchar(255) null,
    curr_content varchar(255) null,
    prev_content_html longtext null,
    curr_content_html longtext null,
    name varchar(255) null,
    email varchar(255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists refresh_token (
    id bigint auto_increment primary key,
    user_id bigint null,
    token varchar(500) null,
    is_active tinyint(1) null,
    expire_in bigint null,
    ip_addr varchar(255) null,
    agent varchar(255) null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_refresh_token_token unique (token)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_refresh_token_user_id on refresh_token (user_id);
create table if not exists repayment_dividend_schedule (
    id bigint auto_increment primary key,
    project_id bigint null,
    project_type varchar(255) null,
    project_share_qty bigint null,
    project_share_amount bigint null,
    decide_date datetime null,
    execute_date datetime null,
    return_type varchar(255) null,
    status varchar(255) not null,
    repayment_amount bigint null,
    repayment_per_share bigint null,
    previous_repayment_reserve bigint null,
    repayment_reserve bigint null,
    dividend_amount bigint null,
    net_dividend_after_broker_fee bigint null,
    broker_fee bigint null,
    return_rate float null,
    dividend_per_share bigint null,
    withholding_tax bigint null,
    investment_balance bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    project_spv_id bigint null,
    num_investors bigint null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists repayment_schedule (
    id bigint auto_increment primary key,
    project_id bigint null,
    repayment_principal bigint null,
    interest_amount bigint null,
    broker_fee bigint null,
    revenue bigint null,
    repayment_date datetime null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_repayment_schedule_project_id on repayment_schedule (project_id);
create table if not exists repayment_settlement_mapping (
    repayment_schedule_id bigint null,
    settlement_schedule_id bigint not null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists report_output (
    id bigint auto_increment primary key,
    user_id bigint null,
    spv_id bigint null,
    is_open tinyint(1) null,
    title varchar(255) not null,
    submission_type varchar(255) not null,
    report_type varchar(255) not null,
    year varchar(255) not null,
    quarter varchar(255) not null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    file_url varchar(255) null,
    at_open datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists role (
    id bigint auto_increment primary key,
    user_id bigint null,
    role_code varchar(100) null,
    created datetime null,
    updated datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists roles_skip_url (
    id bigint auto_increment primary key,
    path varchar(255) not null,
    method varchar(100) null,
    created datetime null,
    updated datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists settlement (
    id bigint auto_increment primary key,
    project_id bigint null,
    project_type varchar(255) null,
    schedule_id bigint not null,
    round int null,
    settlement_type varchar(255) null,
    status varchar(255) not null,
    share_dividend_amount float null,
    share_repayment_amount float null,
    project_share_qty bigint null,
    dividend_amount bigint null,
    interest_amount bigint null,
    repayment_amount bigint null,
    broker_fee bigint null,
    with_holding_tax bigint null,
    etc_profit bigint null,
    repayment_reserve bigint null,
    before_repayment_reserve bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    project_share_amount bigint null,
    before_etc_profit bigint null,
    constraint UQE_settlement_schedule_id unique (schedule_id)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists settlement_schedule (
    id bigint auto_increment primary key,
    project_id bigint null,
    start_date datetime null,
    end_date datetime null,
    dividend_date datetime null,
    is_confirm tinyint(1) null,
    execution_status varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists share (
    id varchar(255) not null primary key,
    project_id bigint null,
    category_type varchar(255) null,
    project_type varchar(255) null,
    owner_id bigint not null,
    update_id bigint not null,
    `desc` varchar(255) null,
    invest_status varchar(255) null,
    amount bigint null,
    qty bigint null,
    price bigint null,
    invest_date datetime not null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    contract_before_page_check_date datetime null,
    contract_after_page_check_date datetime null,
    silent_partnerships_check_date datetime null,
    project_report_check_date datetime null,
    share_type varchar(255) default 'BASIC_SHARE' null,
    commitment_date datetime null,
    constraint UQE_share_id unique (id)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_share_owner_id on share (owner_id);
create index IDX_share_project_id on share (project_id);
create table if not exists share_file (
    id bigint auto_increment primary key,
    share_id varchar(255) null,
    file_name varchar(255) not null,
    original_name varchar(255) not null,
    file_size bigint not null,
    file_extension varchar(255) not null,
    file_path varchar(255) not null,
    file_url varchar(255) not null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_shares_file_file_name unique (file_name)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists sign_up_request (
    id bigint auto_increment primary key,
    request_id varchar(40) not null,
    email varchar(255) null,
    url varchar(255) null,
    email_expire datetime null,
    created datetime null,
    send_count int null,
    url_expire datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists sms_msg (
    id bigint auto_increment primary key,
    msg_id varchar(255) null,
    title varchar(255) null,
    name varchar(255) null,
    sender varchar(255) not null,
    status varchar(255) null,
    sms_type varchar(255) null,
    sms_content text null,
    count bigint null,
    send_date datetime null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists spv (
    id bigint auto_increment primary key,
    name varchar(255) not null,
    email varchar(255) not null,
    address varchar(255) null,
    address2 varchar(255) null,
    city varchar(255) null,
    zip_code varchar(255) null,
    province varchar(255) null,
    corp_no varchar(255) not null,
    phone varchar(255) not null,
    owner_id bigint null,
    update_id bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_spv_corp_no unique (corp_no)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_spv_owner_id on spv (owner_id);
create index IDX_spv_update_id on spv (update_id);
create table if not exists spv_changed_content (
    id bigint auto_increment primary key,
    info_id bigint null,
    content_type varchar(255) null,
    prev_content varchar(255) null,
    curr_content varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists spv_info_history (
    id bigint auto_increment primary key,
    spv_id bigint null,
    substance_type varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists subscription_mail (
    id bigint auto_increment primary key,
    email varchar(255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists transaction_info_history (
    id bigint auto_increment primary key,
    share_id varchar(255) null,
    substance_type varchar(255) null,
    process_method varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    content_type varchar(255) null,
    prev_content varchar(255) null,
    curr_content varchar(255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists transfer (
    id bigint auto_increment primary key,
    item_id int not null,
    request_id bigint not null,
    transaction_id varchar(40) not null,
    user_id bigint null,
    project_id bigint null,
    transfer_date datetime null,
    transfer_amount bigint null,
    account_number varchar(255) null,
    bank_code varchar(255) null,
    bank_name varchar(255) null,
    branch_code varchar(255) null,
    branch_name varchar(255) null,
    depositor varchar(255) null,
    account_type_code varchar(255) null,
    transfer_fee bigint null,
    transaction_detail_type varchar(255) null,
    created datetime null,
    updated datetime null,
    transfer_status varchar(255) null,
    repayment_date datetime null,
    error_code varchar(255) null,
    error_msg varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_transfer_transaction_id on transfer (transaction_id);
create table if not exists transfer_request (
    id bigint auto_increment primary key,
    transfer_date datetime null,
    request_total_amount bigint not null,
    request_total_cnt int not null,
    transfer_fee bigint null,
    error_code varchar(255) null,
    error_msg varchar(255) null,
    request_status varchar(255) null,
    result_code varchar(255) null,
    apply_no varchar(255) null,
    apply_end_datetime varchar(255) null,
    created datetime null,
    updated datetime null,
    completed_transfer varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists trust_info_history (
    id bigint auto_increment primary key,
    print_status varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    period varchar (255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists user (
    id bigint auto_increment primary key,
    uid varchar(255) not null,
    login_type varchar(255) null,
    user_type varchar(255) null,
    common_id varchar(40) not null,
    investor_type varchar(255) null,
    name varchar(255) null,
    email varchar(255) not null,
    password varchar(255) not null,
    salt varchar(255) not null,
    mobile_phone varchar(255) null,
    sign_up_date datetime null,
    login_fail_cnt int null,
    last_login_date datetime null,
    block_invest tinyint(1) null,
    zip_code varchar(255) null,
    address varchar(255) null,
    address2 varchar(255) null,
    province varchar(255) null,
    city varchar(255) not null,
    audit_status varchar(255) null,
    investment_purpose varchar(255) not null,
    is_surplus_fund tinyint(1) not null,
    is_newsletter tinyint(1) not null,
    newsletter_agree_at datetime null,
    is_sms_cert tinyint(1) null,
    sms_cert_at datetime null,
    bank_account varchar(255) null,
    bank_code varchar(255) null,
    branch_code varchar(255) null,
    depositor varchar(255) null,
    bank_name varchar(255) not null,
    branch_name varchar(255) null,
    deposit_type varchar(255) not null,
    postal_code varchar(255) not null,
    postal_code_status tinyint(1) not null,
    postal_code_send_date datetime null,
    sign_up_approval_date datetime null,
    invalid_reason varchar(255) null,
    code_approval_date datetime null,
    is_address_change tinyint(1) null,
    my_number text null,
    agree_pro_investor tinyint(1) not null,
    is_ekyc tinyint(1) null,
    anti_research_date datetime null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    hold_user_check_code varchar(255) null,
    constraint UQE_user_email unique (email),
    constraint UQE_user_uid unique (uid)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists user_account (
    id bigint auto_increment primary key,
    user_id bigint null,
    user_name varchar(255) null,
    balance bigint null,
    created datetime null,
    updated datetime null,
    user_type varchar(255) null,
    version bigint default 1 null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_user_account_user_id on user_account (user_id);
create table if not exists user_antisocial_check (
    id bigint auto_increment primary key,
    check_user_id bigint not null,
    checker varchar(255) null,
    check_date datetime null,
    check_status int null,
    updated datetime null,
    created datetime null,
    invalid_reason varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists user_changed_content (
    id bigint auto_increment primary key,
    info_id bigint null,
    content_type varchar(255) null,
    prev_content varchar(255) null,
    curr_content varchar(255) null,
    prev_content_html longtext null,
    curr_content_html longtext null,
    prev_url varchar(255) null,
    curr_url varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists user_file (
    id bigint auto_increment primary key,
    user_id bigint not null,
    file_name varchar(255) not null,
    original_name varchar(255) not null,
    file_type varchar(255) null,
    file_size bigint not null,
    file_extension varchar(255) not null,
    file_path varchar(255) not null,
    file_url varchar(255) not null,
    owner_id bigint not null,
    update_id bigint null,
    user_type varchar(255) not null,
    etc varchar(255) null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    constraint UQE_user_file_file_name unique (file_name)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists user_info_history (
    id bigint auto_increment primary key,
    user_id bigint null,
    process_user_id bigint null,
    substance_type varchar(255) null,
    process_method varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    created datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists user_notice_mapping (
    msg_id bigint not null,
    user_id bigint not null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists user_quarter_balance (
    id bigint auto_increment primary key,
    user_id bigint null,
    year int null,
    quarter varchar(255) null,
    balance bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists user_roles (
    id bigint auto_increment primary key,
    role_code varchar(100) null,
    role_name varchar(100) null,
    created datetime null,
    updated datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists user_transaction (
    id bigint auto_increment primary key,
    user_id bigint null,
    user_name varchar(255) null,
    project_id bigint null,
    account_number varchar(255) null,
    project_type varchar(255) null,
    amount bigint null,
    balance bigint null,
    transaction_type varchar(255) null,
    transaction_detail_type varchar(255) null,
    transaction_date datetime null,
    created datetime null,
    updated datetime null,
    user_type varchar(255) null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists user_virtual_account (
    id bigint auto_increment primary key,
    user_id bigint null,
    va_bank_name varchar(255) null,
    va_branch_code varchar(255) null,
    va_branch_name varchar(255) null,
    va_account_number varchar(255) null,
    va_holder_name_kana varchar(255) null,
    project_type varchar(255) null,
    linked_status varchar(255) null,
    update_date datetime null,
    created datetime null,
    updated datetime null,
    constraint UQE_user_virtual_account_id unique (id),
    constraint UQE_user_virtual_account_va_account_number unique (va_account_number)
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists va_deposit (
    id bigint auto_increment primary key,
    user_id bigint null,
    va_id varchar(255) null,
    transaction_date varchar(255) null,
    value_date varchar(255) null,
    va_branch_code varchar(255) null,
    va_branch_name_kana varchar(255) null,
    va_account_name_kana varchar(255) null,
    remitter_name_kana varchar(255) null,
    payment_bank_name varchar(255) null,
    payment_branch_name varchar(255) null,
    partner_name varchar(255) null,
    remarks varchar(255) null,
    item_key varchar(255) null,
    va_account_number varchar(255) not null,
    deposit_amount bigint null,
    deposit_date datetime null,
    use_amount bigint null,
    refund_amount bigint null,
    deposit_status varchar(255) null,
    transaction_detail_type varchar(255) null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists withdrawal (
    id bigint auto_increment primary key,
    investor_id varchar(255) not null,
    orderer varchar(255) not null,
    amount bigint null,
    fee bigint null,
    request_date datetime null,
    complete_date datetime null,
    status varchar(255) null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    close_date datetime null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists withdrawal_close (
    id bigint auto_increment primary key,
    closer_id bigint not null,
    close_date datetime not null,
    transfer_date datetime null,
    total bigint default 0 not null,
    fee bigint default 0 not null,
    amount bigint default 0 not null,
    status varchar(30) default 'WAIT' not null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_withdrawal_close_close_date on withdrawal_close (close_date);
create table if not exists withdrawal_close_mapping (
    id bigint auto_increment primary key,
    close_id bigint not null,
    withdrawal_id bigint not null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_withdrawal_close_mapping_close_id on withdrawal_close_mapping (close_id);
create table if not exists withdrawal_execute_mapping (
    id bigint auto_increment primary key,
    history_id bigint null,
    transfer_id bigint null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create table if not exists withdrawal_transfer_mapping (
    id bigint auto_increment primary key,
    transaction_id varchar(40) not null,
    withdrawal_id bigint not null
) charset = utf8mb4 collate = utf8mb4_general_ci;
create index IDX_withdrawal_transfer_mapping_withdrawal_id on withdrawal_transfer_mapping (withdrawal_id);
create or replace definer = admin @`%` view deposit_match_data as
select `userT`.`id` AS `user_id`,
    `userT`.`name` AS `investor_name`,
    `drT`.`investment_amount` AS `investment_amount`,
    `drT`.`deposit_total_amount` AS `deposit_total_amount`,
    if(
        (
            `drT`.`investment_amount` > `drT`.`deposit_total_amount`
        ),
        'LACK_DEPOSIT',
        'SAME_DEPOSIT'
    ) AS `deposit_match_status`,
    (
        case
            when (
                (
                    `drT`.`investment_amount` - `drT`.`deposit_total_amount`
                ) < 0
            ) then 0
            else (
                `drT`.`investment_amount` - `drT`.`deposit_total_amount`
            )
        end
    ) AS `deposit_match_amount`,
    (
        case
            when (
                (
                    `drT`.`deposit_total_amount` - `drT`.`investment_amount`
                ) > 0
            ) then 0
            else (
                `drT`.`deposit_total_amount` - `drT`.`investment_amount`
            )
        end
    ) AS `lack_amount`,
    `drT`.`update_date` AS `update_date`
from (
        (
            (
                select `energysharesdb`.`user`.`id` AS `id`,
                    `energysharesdb`.`user`.`uid` AS `uid`,
                    `energysharesdb`.`user`.`login_type` AS `login_type`,
                    `energysharesdb`.`user`.`user_type` AS `user_type`,
                    `energysharesdb`.`user`.`common_id` AS `common_id`,
                    `energysharesdb`.`user`.`investor_type` AS `investor_type`,
                    `energysharesdb`.`user`.`name` AS `name`,
                    `energysharesdb`.`user`.`email` AS `email`,
                    `energysharesdb`.`user`.`password` AS `password`,
                    `energysharesdb`.`user`.`salt` AS `salt`,
                    `energysharesdb`.`user`.`mobile_phone` AS `mobile_phone`,
                    `energysharesdb`.`user`.`sign_up_date` AS `sign_up_date`,
                    `energysharesdb`.`user`.`login_fail_cnt` AS `login_fail_cnt`,
                    `energysharesdb`.`user`.`last_login_date` AS `last_login_date`,
                    `energysharesdb`.`user`.`block_invest` AS `block_invest`,
                    `energysharesdb`.`user`.`zip_code` AS `zip_code`,
                    `energysharesdb`.`user`.`address` AS `address`,
                    `energysharesdb`.`user`.`address2` AS `address2`,
                    `energysharesdb`.`user`.`province` AS `province`,
                    `energysharesdb`.`user`.`city` AS `city`,
                    `energysharesdb`.`user`.`audit_status` AS `audit_status`,
                    `energysharesdb`.`user`.`investment_purpose` AS `investment_purpose`,
                    `energysharesdb`.`user`.`is_surplus_fund` AS `is_surplus_fund`,
                    `energysharesdb`.`user`.`is_newsletter` AS `is_newsletter`,
                    `energysharesdb`.`user`.`newsletter_agree_at` AS `newsletter_agree_at`,
                    `energysharesdb`.`user`.`is_sms_cert` AS `is_sms_cert`,
                    `energysharesdb`.`user`.`sms_cert_at` AS `sms_cert_at`,
                    `energysharesdb`.`user`.`bank_account` AS `bank_account`,
                    `energysharesdb`.`user`.`bank_code` AS `bank_code`,
                    `energysharesdb`.`user`.`branch_code` AS `branch_code`,
                    `energysharesdb`.`user`.`depositor` AS `depositor`,
                    `energysharesdb`.`user`.`bank_name` AS `bank_name`,
                    `energysharesdb`.`user`.`branch_name` AS `branch_name`,
                    `energysharesdb`.`user`.`deposit_type` AS `deposit_type`,
                    `energysharesdb`.`user`.`postal_code` AS `postal_code`,
                    `energysharesdb`.`user`.`postal_code_status` AS `postal_code_status`,
                    `energysharesdb`.`user`.`postal_code_send_date` AS `postal_code_send_date`,
                    `energysharesdb`.`user`.`sign_up_approval_date` AS `sign_up_approval_date`,
                    `energysharesdb`.`user`.`invalid_reason` AS `invalid_reason`,
                    `energysharesdb`.`user`.`code_approval_date` AS `code_approval_date`,
                    `energysharesdb`.`user`.`is_address_change` AS `is_address_change`,
                    `energysharesdb`.`user`.`my_number` AS `my_number`,
                    `energysharesdb`.`user`.`agree_pro_investor` AS `agree_pro_investor`,
                    `energysharesdb`.`user`.`anti_research_date` AS `anti_research_date`,
                    `energysharesdb`.`user`.`created` AS `created`,
                    `energysharesdb`.`user`.`updated` AS `updated`,
                    `energysharesdb`.`user`.`deleted` AS `deleted`
                from `energysharesdb`.`user`
                where (`energysharesdb`.`user`.`user_type` <> 'ADMIN')
            )
        ) `userT`
        left join (
            select `energysharesdb`.`deposit_request`.`user_id` AS `user_id`,
                sum(
                    `energysharesdb`.`deposit_request`.`request_amount`
                ) AS `investment_amount`,
                sum(
                    `energysharesdb`.`deposit_request`.`deposit_amount`
                ) AS `deposit_total_amount`,
                max(`energysharesdb`.`deposit_request`.`update_date`) AS `update_date`
            from `energysharesdb`.`deposit_request`
            where (
                    `energysharesdb`.`deposit_request`.`deposit_check_status` <> 'CANCEL_DEPOSIT'
                )
            group by `energysharesdb`.`deposit_request`.`user_id`
        ) `drT` on ((`userT`.`id` = `drT`.`user_id`))
    );
create or replace definer = admin @`%` view project_dividend_detail as
select `energysharesdb`.`dividend`.`id` AS `id`,
    `energysharesdb`.`dividend`.`user_id` AS `user_id`,
    `energysharesdb`.`dividend`.`project_id` AS `project_id`,
    `energysharesdb`.`dividend`.`project_type` AS `project_type`,
    `energysharesdb`.`dividend`.`settlement_id` AS `settlement_id`,
    `energysharesdb`.`dividend`.`investment_amount` AS `investment_amount`,
    `energysharesdb`.`dividend`.`investment_qty` AS `investment_qty`,
    `energysharesdb`.`dividend`.`amount` AS `amount`,
    `energysharesdb`.`dividend`.`qty` AS `qty`,
    `energysharesdb`.`dividend`.`deposit_amount` AS `deposit_amount`,
    `energysharesdb`.`dividend`.`tax` AS `tax`,
    `energysharesdb`.`dividend`.`dividend_date` AS `dividend_date`,
    `energysharesdb`.`dividend`.`transaction_status` AS `transaction_status`,
    `energysharesdb`.`dividend`.`transaction_type` AS `transaction_type`,
    `energysharesdb`.`dividend`.`transaction_date` AS `transaction_date`,
    `energysharesdb`.`dividend`.`transaction_id` AS `transaction_id`,
    `energysharesdb`.`dividend`.`created` AS `created`,
    `energysharesdb`.`dividend`.`updated` AS `updated`,
    `energysharesdb`.`dividend`.`deleted` AS `deleted`,
    `energysharesdb`.`project`.`name` AS `project_name`,
    `energysharesdb`.`project`.`project_status` AS `project_status`,
    sum(
        (
            case
                when (
                    `energysharesdb`.`dividend`.`transaction_type` = 'DEPOSIT_REPAYMENT'
                ) then `energysharesdb`.`dividend`.`amount`
                else 0
            end
        )
    ) AS `repayment_amount`,
    sum(
        (
            case
                when (
                    `energysharesdb`.`dividend`.`transaction_type` = 'DEPOSIT_DIVIDEND'
                ) then `energysharesdb`.`dividend`.`amount`
                else 0
            end
        )
    ) AS `dividend_amount`,
    sum(`energysharesdb`.`dividend`.`amount`) AS `total_amount`,
    sum(
        (
            case
                when (
                    `energysharesdb`.`dividend`.`transaction_type` = 'DEPOSIT_DIVIDEND'
                ) then `energysharesdb`.`dividend`.`deposit_amount`
                else 0
            end
        )
    ) AS `total_revenue_amount`,
    sum(`energysharesdb`.`dividend`.`tax`) AS `total_tax`
from (
        (
            (
                (
                    select `energysharesdb`.`settlement`.`id` AS `id`,
                        `energysharesdb`.`settlement`.`project_id` AS `project_id`,
                        `energysharesdb`.`settlement`.`project_type` AS `project_type`,
                        `energysharesdb`.`settlement`.`schedule_id` AS `schedule_id`,
                        `energysharesdb`.`settlement`.`round` AS `round`,
                        `energysharesdb`.`settlement`.`settlement_type` AS `settlement_type`,
                        `energysharesdb`.`settlement`.`status` AS `status`,
                        `energysharesdb`.`settlement`.`share_dividend_amount` AS `share_dividend_amount`,
                        `energysharesdb`.`settlement`.`share_repayment_amount` AS `share_repayment_amount`,
                        `energysharesdb`.`settlement`.`project_share_qty` AS `project_share_qty`,
                        `energysharesdb`.`settlement`.`dividend_amount` AS `dividend_amount`,
                        `energysharesdb`.`settlement`.`interest_amount` AS `interest_amount`,
                        `energysharesdb`.`settlement`.`repayment_amount` AS `repayment_amount`,
                        `energysharesdb`.`settlement`.`broker_fee` AS `broker_fee`,
                        `energysharesdb`.`settlement`.`with_holding_tax` AS `with_holding_tax`,
                        `energysharesdb`.`settlement`.`etc_profit` AS `etc_profit`,
                        `energysharesdb`.`settlement`.`repayment_reserve` AS `repayment_reserve`,
                        `energysharesdb`.`settlement`.`before_repayment_reserve` AS `before_repayment_reserve`,
                        `energysharesdb`.`settlement`.`created` AS `created`,
                        `energysharesdb`.`settlement`.`updated` AS `updated`,
                        `energysharesdb`.`settlement`.`deleted` AS `deleted`
                    from `energysharesdb`.`settlement`
                    where (
                            `energysharesdb`.`settlement`.`status` = 'COMPLETED'
                        )
                )
            ) `settle`
            join `energysharesdb`.`dividend` on (
                (
                    (
                        `settle`.`project_id` = `energysharesdb`.`dividend`.`project_id`
                    )
                    and (
                        `settle`.`id` = `energysharesdb`.`dividend`.`settlement_id`
                    )
                )
            )
        )
        join `energysharesdb`.`project` on (
            (
                `settle`.`project_id` = `energysharesdb`.`project`.`id`
            )
        )
    )
group by `energysharesdb`.`dividend`.`user_id`,
    `energysharesdb`.`dividend`.`project_id`;
create or replace definer = admin @`%` view user_asset_view as
select `u`.`id` AS `user_id`,
    coalesce(`ua`.`balance`, 0) AS `deposit`,
    coalesce(`ioa`.`invest_operation_amount`, 0) AS `invest_operation_amount`,
    coalesce(`isa`.`invest_success_amount`, 0) AS `invest_success_amount`,
    coalesce(`ira`.`invest_request_amount`, 0) AS `invest_request_amount`,
    coalesce(`wra`.`withdrawal_request_amount`, 0) AS `withdrawal_request_amount`,
    coalesce(`uaa`.`unpaid_amount`, 0) AS `unpaid_amount`
from (
        (
            (
                (
                    (
                        (
                            `energysharesdb`.`user` `u`
                            left join `energysharesdb`.`user_account` `ua` on ((`u`.`id` = `ua`.`user_id`))
                        )
                        left join (
                            select `energysharesdb`.`share`.`owner_id` AS `owner_id`,
                                coalesce(sum(`energysharesdb`.`share`.`amount`), 0) AS `invest_operation_amount`
                            from `energysharesdb`.`share`
                            where (
                                    `energysharesdb`.`share`.`invest_status` = 'OPERATION_INVEST'
                                )
                            group by `energysharesdb`.`share`.`owner_id`
                        ) `ioa` on ((`u`.`id` = `ioa`.`owner_id`))
                    )
                    left join (
                        select `energysharesdb`.`share`.`owner_id` AS `owner_id`,
                            coalesce(sum(`energysharesdb`.`share`.`amount`), 0) AS `invest_success_amount`
                        from `energysharesdb`.`share`
                        where (
                                `energysharesdb`.`share`.`invest_status` = 'SUCCESS_INVEST'
                            )
                        group by `energysharesdb`.`share`.`owner_id`
                    ) `isa` on ((`u`.`id` = `isa`.`owner_id`))
                )
                left join (
                    select `s`.`owner_id` AS `owner_id`,
                        coalesce(sum(`d`.`deposit_amount`), 0) AS `invest_request_amount`
                    from (
                            `energysharesdb`.`share` `s`
                            join `energysharesdb`.`deposit_request` `d` on ((`s`.`id` = `d`.`share_id`))
                        )
                    where (`s`.`invest_status` = 'REQUEST_INVEST')
                    group by `s`.`owner_id`
                ) `ira` on ((`u`.`id` = `ira`.`owner_id`))
            )
            left join (
                select `energysharesdb`.`withdrawal`.`investor_id` AS `investor_id`,
                    coalesce(sum(`energysharesdb`.`withdrawal`.`amount`), 0) AS `withdrawal_request_amount`
                from `energysharesdb`.`withdrawal`
                where (
                        `energysharesdb`.`withdrawal`.`status` in ('INCOMPLETE', 'WAIT', 'FAIL')
                    )
                group by `energysharesdb`.`withdrawal`.`investor_id`
            ) `wra` on ((`u`.`id` = `wra`.`investor_id`))
        )
        left join (
            select `energysharesdb`.`deposit_request`.`user_id` AS `user_id`,
                coalesce(
                    sum(
                        (
                            `energysharesdb`.`deposit_request`.`request_amount` - `energysharesdb`.`deposit_request`.`deposit_amount`
                        )
                    ),
                    0
                ) AS `unpaid_amount`
            from `energysharesdb`.`deposit_request`
            where (
                    (
                        `energysharesdb`.`deposit_request`.`deposit_check_status` = 'NO_DEPOSIT'
                    )
                    and (
                        `energysharesdb`.`deposit_request`.`expire_deposit_date` >= now()
                    )
                )
            group by `energysharesdb`.`deposit_request`.`user_id`
        ) `uaa` on ((`u`.`id` = `uaa`.`user_id`))
    );