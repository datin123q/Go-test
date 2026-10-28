create table if not exists admin_profile (
    id bigint auto_increment primary key,
    user_id bigint null,
    validity_start_date datetime null,
    validity_end_date datetime null,
    role_updated datetime null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) collate = utf8mb4_general_ci auto_increment = 19;
create table if not exists answer (
    id bigint auto_increment primary key,
    qna_id bigint not null,
    common_id varchar(40) not null,
    answer varchar(255) null,
    created datetime null,
    updated datetime null
) collate = utf8mb4_general_ci auto_increment = 24;
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
) collate = utf8mb4_general_ci auto_increment = 77;
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
) collate = utf8mb4_general_ci auto_increment = 42;
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
) collate = utf8mb4_general_ci;
create index IDX_audit_status_history_owner_id on audit_status_history (owner_id);
create index IDX_audit_status_history_update_id on audit_status_history (update_id);
create table if not exists casbin_rule (
    p_type varchar(100) null,
    v0 varchar(100) null,
    v1 varchar(100) null,
    v2 varchar(100) null,
    v3 varchar(100) null,
    v4 varchar(100) null,
    v5 varchar(100) null
) collate = utf8mb4_general_ci;
create index IDX_casbin_rule_p_type on casbin_rule (p_type);
create index IDX_casbin_rule_v0 on casbin_rule (v0);
create index IDX_casbin_rule_v1 on casbin_rule (v1);
create index IDX_casbin_rule_v2 on casbin_rule (v2);
create index IDX_casbin_rule_v3 on casbin_rule (v3);
create index IDX_casbin_rule_v4 on casbin_rule (v4);
create index IDX_casbin_rule_v5 on casbin_rule (v5);
create table if not exists consult_note (
    id bigint auto_increment primary key,
    user_id bigint null,
    owner_id bigint null,
    update_id bigint null,
    creator varchar(255) not null,
    request varchar(255) not null,
    memo varchar(255) not null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) collate = utf8mb4_general_ci auto_increment = 9;
create index IDX_consult_note_owner_id on consult_note (owner_id);
create index IDX_consult_note_update_id on consult_note (update_id);
create table if not exists contact (
    id bigint auto_increment primary key,
    name varchar(255) null,
    email varchar(255) null,
    title varchar(255) null,
    contents longtext null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) collate = utf8mb4_general_ci auto_increment = 15;
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
) collate = utf8mb4_general_ci auto_increment = 6;
create index IDX_content_image_resource_id on content_image (resource_id);
create table if not exists contract (
    id bigint auto_increment primary key,
    share_id varchar(255) null,
    signed tinyint(1) null,
    created datetime null,
    deleted datetime null,
    share_qty bigint null,
    share_amount bigint null,
    status varchar(255) null,
    project_id bigint null,
    user_name varchar(255) null,
    contract_after_page_url varchar(255) null,
    updated datetime null
) collate = utf8mb4_general_ci auto_increment = 733;
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
) collate = utf8mb4_general_ci auto_increment = 96;
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
) collate = utf8mb4_general_ci auto_increment = 29;
create table if not exists corporate_sign_up_request (
    id bigint auto_increment primary key,
    request_id varchar(40) not null,
    email varchar(255) null,
    url varchar(255) null,
    expire datetime null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 49;
create table if not exists deposit_mapping (
    id bigint auto_increment primary key,
    request_id bigint null,
    va_deposit_id bigint null
) collate = utf8mb4_general_ci auto_increment = 21;
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
) collate = utf8mb4_general_ci auto_increment = 1098;
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
) collate = utf8mb4_general_ci auto_increment = 572;
create table if not exists deposit_withdrawal_execute_history (
    id bigint auto_increment primary key,
    investor_id bigint null,
    investor_name varchar(255) null,
    account_number varchar(255) null,
    amount bigint null,
    operator_name varchar(255) null,
    operator_email varchar(255) null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 12;
create table if not exists display_code (
    id bigint auto_increment primary key,
    code varchar(255) null,
    active tinyint(1) null,
    created datetime null,
    updated datetime null
) collate = utf8mb4_general_ci auto_increment = 5;
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
    deleted datetime null
) collate = utf8mb4_general_ci auto_increment = 116;
create index IDX_dividend_transaction_id on dividend (transaction_id);
create table if not exists ekyc_request (
    id bigint auto_increment primary key,
    request_id varchar(255) not null,
    url varchar(255) null,
    expire datetime null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 186;
create table if not exists email_code_request (
    id bigint auto_increment primary key,
    email varchar(255) null,
    code varchar(255) null,
    expire datetime null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 11;
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
) collate = utf8mb4_general_ci auto_increment = 17;
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
    content longtext not null,
    equipment_name varchar(255) null,
    equipment_zip_code varchar(255) null,
    equipment_province varchar(255) null,
    equipment_city varchar(255) null,
    equipment_address varchar(255) null,
    equipment_address2 varchar(255) null,
    equipment_id varchar(255) null,
    develop_scale varchar(255) null,
    equipment_date datetime null,
    fixed_purchase_price varchar(255) null,
    fixed_purchase_period_date datetime null,
    purchase_period varchar(255) null,
    application_rule varchar(255) null,
    solar_total_size varchar(255) null,
    system_access_power_company varchar(255) null,
    process varchar(255) null,
    schedule_date datetime null,
    output_suppression varchar(255) null,
    payment_status varchar(255) null,
    distance varchar(255) null,
    land_status longtext null,
    facility_status longtext null,
    grid_connection_status longtext null,
    license_status longtext null,
    is_forest_development tinyint(1) null,
    is_environmental_impact tinyint(1) null,
    request_status varchar(255) null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) collate = utf8mb4_general_ci auto_increment = 42;
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
) collate = utf8mb4_general_ci auto_increment = 121;
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
) collate = utf8mb4_general_ci auto_increment = 2;
create table if not exists holiday (
    id bigint auto_increment primary key,
    year int null,
    month int null,
    day int null,
    date varchar(255) null,
    created datetime null,
    holiday_date datetime null
) collate = utf8mb4_general_ci auto_increment = 129;
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
) collate = utf8mb4_general_ci auto_increment = 53;
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
) collate = utf8mb4_general_ci auto_increment = 242;
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
) collate = utf8mb4_general_ci auto_increment = 219;
create table if not exists japan_bank (
    id bigint auto_increment primary key,
    bank_code varchar(255) not null,
    bank_name_kana varchar(255) not null,
    bank_name_kanji varchar(255) not null,
    member varchar(255) not null,
    created datetime null,
    updated datetime null
) collate = utf8mb4_general_ci auto_increment = 2491;
create table if not exists japan_bank_branch (
    id bigint auto_increment primary key,
    bank_code varchar(255) not null,
    branch_code varchar(255) not null,
    branch_name_kana varchar(255) not null,
    branch_name_kanji varchar(255) not null,
    postal_code varchar(255) not null,
    branch_address varchar(255) not null,
    phone_number varchar(255) not null,
    clearing_house_number varchar(255) not null,
    sort_code varchar(255) not null,
    member varchar(255) not null,
    created datetime null,
    updated datetime null
) collate = utf8mb4_general_ci auto_increment = 64318;
create table if not exists job_execute_time (
    id bigint auto_increment primary key,
    deposit_check_time datetime null,
    va_account_issue_time datetime null,
    transfer_request_time datetime null,
    transfer_result_time datetime null,
    deposit_process_time datetime null
) collate = utf8mb4_general_ci auto_increment = 2;
create table if not exists jwt_skip_url (
    id bigint auto_increment primary key,
    path varchar(255) not null,
    method varchar(100) null,
    created datetime null,
    updated datetime null
) collate = utf8mb4_general_ci auto_increment = 67;
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
) collate = utf8mb4_general_ci auto_increment = 2808;
create index IDX_login_history_user_id on login_history (user_id);
create table if not exists mail (
    id bigint auto_increment primary key,
    status varchar(255) null,
    template_type varchar(255) null,
    subject varchar(255) null,
    name varchar(255) null,
    sender varchar(255) null,
    receiver longtext null,
    content longtext null,
    mail_type varchar(255) null,
    count bigint null,
    send_date datetime null,
    updated datetime null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 5543;
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
) collate = utf8mb4_general_ci auto_increment = 9885;
create index IDX_mail_request_mail_id on mail_request (mail_id);
create table if not exists mail_template (
    id bigint auto_increment primary key,
    template_name varchar(255) null,
    subject varchar(255) null,
    content longtext null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 54;
create table if not exists mobile_phone_verification_request (
    id bigint auto_increment primary key,
    request_id varchar(40) not null,
    user_id bigint null,
    code varchar(255) null,
    mobile_phone varchar(255) null,
    status varchar(255) null,
    expire datetime null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 19;
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
) collate = utf8mb4_general_ci auto_increment = 38;
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
) collate = utf8mb4_general_ci auto_increment = 23;
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
) collate = utf8mb4_general_ci auto_increment = 35;
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
) collate = utf8mb4_general_ci auto_increment = 23751;
create table if not exists password_reset_request (
    id bigint auto_increment primary key,
    request_id varchar(40) not null,
    email varchar(255) null,
    url varchar(255) null,
    expire datetime null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 24;
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
) collate = utf8mb4_general_ci auto_increment = 17;
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
) collate = utf8mb4_general_ci auto_increment = 49;
create table if not exists popular_info (
    id bigint auto_increment primary key,
    `index` int null,
    project_id bigint null,
    posted tinyint(1) null,
    project_name varchar(255) null,
    updated datetime null,
    created datetime null,
    constraint UQE_popular_info_index unique (`index`)
) collate = utf8mb4_general_ci auto_increment = 37;
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
) collate = utf8mb4_general_ci auto_increment = 490;
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
    equipment_name varchar(255) null,
    equipment_zip_code varchar(255) null,
    equipment_province varchar(255) null,
    equipment_city varchar(255) null,
    equipment_address varchar(255) null,
    equipment_address2 varchar(255) null,
    equipment_id varchar(255) null,
    develop_scale varchar(255) null,
    equipment_date datetime null,
    fixed_purchase_price varchar(255) null,
    fixed_purchase_period_date datetime null,
    purchase_period varchar(255) null,
    application_rule varchar(255) null,
    solar_total_size varchar(255) null,
    system_access_power_company varchar(255) null,
    process varchar(255) null,
    schedule_date datetime null,
    output_suppression varchar(255) null,
    payment_status varchar(255) null,
    distance varchar(255) null,
    is_forest_development tinyint(1) null,
    is_environmental_impact tinyint(1) null,
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
    contract_before_page_2_date datetime null,
    is_achieve_minimum_fundraising tinyint(1) default 0 not null,
    is_amount_limit tinyint(1) default 0 null,
    total_evaluation_score float null,
    about_evaluation longtext null
) auto_increment = 180;
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
) collate = utf8mb4_general_ci auto_increment = 81;
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
) collate = utf8mb4_general_ci;
create table if not exists project_changed_content (
    id bigint auto_increment primary key,
    info_id bigint null,
    content_type varchar(255) null,
    prev_content varchar(255) null,
    curr_content varchar(255) null,
    prev_content_html longtext null,
    curr_content_html longtext null,
    prev_url varchar(255) null,
    curr_url varchar(255) null
) collate = utf8mb4_general_ci auto_increment = 1419;
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
) collate = utf8mb4_general_ci auto_increment = 558;
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
) collate = utf8mb4_general_ci auto_increment = 320;
create table if not exists project_info_history (
    id bigint auto_increment primary key,
    project_id bigint null,
    substance_type varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 1389;
create table if not exists project_loan_execute (
    id bigint auto_increment primary key,
    project_id bigint null,
    loan_execute_status varchar(255) null,
    loan_amount bigint null,
    loan_execute_date datetime null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 33;
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
) collate = utf8mb4_general_ci auto_increment = 78;
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
) collate = utf8mb4_general_ci;
create index IDX_project_request_user_id on project_request (user_id);
create table if not exists project_view_count (
    id bigint auto_increment primary key,
    project_id bigint null,
    view_count bigint null,
    updated datetime null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 177;
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
) collate = utf8mb4_general_ci auto_increment = 56;
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
) collate = utf8mb4_general_ci auto_increment = 29;
create table if not exists refresh_token (
    i_d bigint auto_increment primary key,
    user_id bigint null,
    token varchar(500) null,
    is_active tinyint(1) null,
    expire_in bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    ip_addr varchar(255) null,
    agent varchar(255) null,
    constraint UQE_refresh_token_token unique (token)
) collate = utf8mb4_general_ci auto_increment = 2808;
create index IDX_refresh_token_user_id on refresh_token (user_id);
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
) collate = utf8mb4_general_ci auto_increment = 78;
create index IDX_repayment_schedule_project_id on repayment_schedule (project_id);
create table if not exists repayment_settlement_mapping (
    repayment_schedule_id bigint null,
    settlement_schedule_id bigint not null,
    created datetime null
) collate = utf8mb4_general_ci;
create table if not exists report_file (
    id bigint auto_increment primary key,
    file_name varchar(255) not null,
    original_name varchar(255) not null,
    file_size bigint not null,
    file_extension varchar(255) not null,
    file_path varchar(255) not null,
    file_url varchar(255) not null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) collate = utf8mb4_general_ci auto_increment = 1608;
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
    report_file_id bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) collate = utf8mb4_general_ci auto_increment = 1584;
create table if not exists report_output_group (
    id bigint auto_increment primary key,
    is_open tinyint(1) null,
    submission_type varchar(255) not null,
    report_type varchar(255) not null,
    year varchar(255) not null,
    quarter varchar(255) not null,
    report_file_id bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) collate = utf8mb4_general_ci auto_increment = 25;
create table if not exists role (
    id bigint auto_increment primary key,
    user_id bigint null,
    role_code varchar(100) null,
    created datetime null,
    updated datetime null
) collate = utf8mb4_general_ci auto_increment = 796;
create index role_user_id_index on role (user_id);
create table if not exists roles_skip_url (
    id bigint auto_increment primary key,
    path varchar(255) not null,
    method varchar(100) null,
    created datetime null,
    updated datetime null
) collate = utf8mb4_general_ci auto_increment = 67;
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
) collate = utf8mb4_general_ci auto_increment = 57;
create table if not exists settlement_schedule (
    id bigint auto_increment primary key,
    project_id bigint null,
    start_date datetime null,
    end_date datetime null,
    dividend_date datetime null,
    is_confirm tinyint(1) null,
    execution_status varchar(255) null
) collate = utf8mb4_general_ci auto_increment = 60;
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
    contract_before_page2_check_date datetime null,
    constraint UQE_share_id unique (id)
) collate = utf8mb4_general_ci;
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
) collate = utf8mb4_general_ci auto_increment = 751;
create table if not exists sign_up_request (
    id bigint auto_increment primary key,
    request_id varchar(40) not null,
    email varchar(255) null,
    url varchar(255) null,
    expire datetime null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 275;
create table if not exists sms_msg (
    id bigint auto_increment primary key,
    msg_id varchar(255) null,
    title varchar(255) null,
    name varchar(255) null,
    sender varchar(255) not null,
    receiver varchar(255) null,
    status varchar(255) null,
    sms_type varchar(255) null,
    sms_content text null,
    count bigint null,
    send_date datetime null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 36;
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
) collate = utf8mb4_general_ci auto_increment = 32;
create index IDX_spv_owner_id on spv (owner_id);
create index IDX_spv_update_id on spv (update_id);
create table if not exists spv_changed_content (
    id bigint auto_increment primary key,
    info_id bigint null,
    content_type varchar(255) null,
    prev_content varchar(255) null,
    curr_content varchar(255) null
) collate = utf8mb4_general_ci auto_increment = 21;
create table if not exists spv_info_history (
    id bigint auto_increment primary key,
    spv_id bigint null,
    substance_type varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 7;
create table if not exists subscription_mail (
    id bigint auto_increment primary key,
    email varchar(255) null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 14;
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
) collate = utf8mb4_general_ci auto_increment = 1232;
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
    updated datetime null
) collate = utf8mb4_general_ci auto_increment = 29;
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
    updated datetime null
) collate = utf8mb4_general_ci auto_increment = 25;
create table if not exists trust_info_history (
    id bigint auto_increment primary key,
    print_status varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    `period` varchar(255) null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 37;
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
    is_online_certification tinyint(1) null,
    anti_research_date datetime null,
    created datetime null,
    updated datetime null,
    deleted datetime null,
    hold_user_check_code varchar(255) null,
    constraint UQE_user_email unique (email),
    constraint UQE_user_uid unique (uid)
) collate = utf8mb4_general_ci auto_increment = 10478;
create table if not exists user_account (
    id bigint auto_increment primary key,
    user_id bigint null,
    user_name varchar(255) null,
    balance bigint null,
    created datetime null,
    updated datetime null,
    user_type varchar(255) null,
    version bigint default 1 null
) collate = utf8mb4_general_ci auto_increment = 514;
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
) collate = utf8mb4_general_ci auto_increment = 111;
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
) collate = utf8mb4_general_ci auto_increment = 1527;
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
) collate = utf8mb4_general_ci auto_increment = 324;
create index user_file_user_id_index on user_file (user_id);
create table if not exists user_info_history (
    id bigint auto_increment primary key,
    user_id bigint null,
    process_user_id bigint null,
    substance_type varchar(255) null,
    process_method varchar(255) null,
    name varchar(255) null,
    email varchar(255) null,
    created datetime null
) collate = utf8mb4_general_ci auto_increment = 966;
create table if not exists user_notice_mapping (
    msg_id bigint not null,
    user_id bigint not null
) collate = utf8mb4_general_ci;
create table if not exists user_quarter_balance (
    id bigint auto_increment primary key,
    user_id bigint null,
    year int null,
    quarter varchar(255) null,
    balance bigint null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) collate = utf8mb4_general_ci auto_increment = 947;
create table if not exists user_roles (
    id bigint auto_increment primary key,
    role_code varchar(100) null,
    role_name varchar(100) null,
    created datetime null,
    updated datetime null
) collate = utf8mb4_general_ci auto_increment = 12;
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
) collate = utf8mb4_general_ci auto_increment = 1513;
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
) collate = utf8mb4_general_ci auto_increment = 90803;
create index user_virtual_account_user_id on user_virtual_account (user_id);
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
) collate = utf8mb4_general_ci auto_increment = 677;
create table if not exists withdrawal (
    id bigint auto_increment primary key,
    investor_id varchar(255) not null,
    orderer varchar(255) not null,
    amount bigint null,
    fee bigint null,
    amount_sum bigint null,
    request_date datetime null,
    complete_date datetime null,
    status tinyint(1) null,
    created datetime null,
    updated datetime null,
    deleted datetime null
) collate = utf8mb4_general_ci auto_increment = 86;
create table if not exists withdrawal_execute_mapping (
    id bigint auto_increment primary key,
    history_id bigint null,
    transfer_id bigint null
) collate = utf8mb4_general_ci auto_increment = 12;
#----------------------------v1.0.7------------------------------
alter table project
add project_progress_percent int null;
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
) charset = utf8mb4;
create index IDX_project_milestone_milestone_no on project_milestone (milestone_no);
create index IDX_project_milestone_milestone_status on project_milestone (milestone_status);
create index IDX_project_milestone_project_id on project_milestone (project_id);
#----------------------------v1.0.9------------------------------
alter table funding_request drop column content;
alter table funding_request drop column equipment_zip_code;
alter table funding_request drop column equipment_province;
alter table funding_request drop column equipment_city;
alter table funding_request drop column equipment_id;
alter table funding_request drop column develop_scale;
alter table funding_request drop column equipment_date;
alter table funding_request drop column fixed_purchase_price;
alter table funding_request drop column fixed_purchase_period_date;
alter table funding_request drop column purchase_period;
alter table funding_request drop column application_rule;
alter table funding_request drop column solar_total_size;
alter table funding_request drop column system_access_power_company;
alter table funding_request drop column process;
alter table funding_request drop column schedule_date;
alter table funding_request drop column output_suppression;
alter table funding_request drop column payment_status;
alter table funding_request drop column distance;
alter table funding_request drop column land_status;
alter table funding_request drop column facility_status;
alter table funding_request drop column grid_connection_status;
alter table funding_request drop column license_status;
alter table funding_request drop column is_forest_development;
alter table funding_request drop column is_environmental_impact;
alter table project
add is_visible_investment_info tinyint default 0 null;
alter table project drop column equipment_name;
alter table project drop column equipment_zip_code;
alter table project drop column equipment_province;
alter table project drop column equipment_city;
alter table project drop column equipment_address;
alter table project drop column equipment_address2;
alter table project drop column equipment_id;
alter table project drop column develop_scale;
alter table project drop column equipment_date;
alter table project drop column fixed_purchase_price;
alter table project drop column fixed_purchase_period_date;
alter table project drop column purchase_period;
alter table project drop column application_rule;
alter table project drop column solar_total_size;
alter table project drop column system_access_power_company;
alter table project drop column process;
alter table project drop column schedule_date;
alter table project drop column output_suppression;
alter table project drop column payment_status;
alter table project drop column distance;
alter table project drop column is_forest_development;
alter table project drop column is_environmental_impact;
#----------------------------v1.0.8------------------------------
#----------------------------v1.0.9------------------------------
alter table refresh_token change i_d id bigint auto_increment;
create index IDX_mail_template_template_name on mail_template (template_name);
#----------------------------v1.0.13------------------------------
#----------------------------v1.0.16------------------------------
create table marketing_code (
    id bigint auto_increment primary key,
    code varchar(255) not null,
    campaign varchar(255) not null,
    sign_up_request_id varchar(255) not null,
    email varchar(255) not null,
    created datetime null,
    updated datetime null
) collate = utf8mb4_general_ci;
alter table email_code_request
add send_count int null;
alter table mobile_phone_verification_request
add send_count int null;
#-----------------------v1.0.17-----------------------------------------
alter table sign_up_request
add column send_count int null;
alter table corporate_sign_up_request
add send_count int null;
#------------------------v1.0.18----------------------------------------
alter table marketing_code
add access_key varchar(255);
#------------------------v1.0.19----------------------------------------
alter table email_code_request
add first_req tinyint(1) null;
alter table email_code_request
add user_id bigint null;
alter table email_code_request drop column send_count;
alter table mobile_phone_verification_request
add first_req tinyint(1) null;
alter table mobile_phone_verification_request drop column send_count;
#------------------------v1.0.20----------------------------------------
alter table mobile_phone_verification_request
add phone_type varchar(255) null;
#------------------------v1.0.21----------------------------------------
alter table withdrawal
modify status VARCHAR(30) null;
update withdrawal
set status = case
        when status = 0 then 'INCOMPLETE'
        when status = 1 then 'COMPLETE'
    end;
alter table withdrawal
add close_date datetime null;
alter table withdrawal drop column amount_sum;
create table if not exists withdrawal_close (
    id bigint auto_increment primary key,
    closer_id bigint not null,
    close_date datetime not null,
    transfer_date datetime null,
    total bigint not null default 0,
    fee bigint not null default 0,
    amount bigint not null default 0,
    status varchar(30) not null default 'WAIT'
) collate = utf8mb4_general_ci;
create index IDX_withdrawal_close_close_date on withdrawal_close (close_date);
create table if not exists withdrawal_close_mapping (
    id bigint auto_increment primary key,
    close_id bigint not null,
    withdrawal_id bigint not null
) collate = utf8mb4_general_ci;
create index IDX_withdrawal_close_mapping_close_id on withdrawal_close_mapping (close_id);
create table if not exists withdrawal_transfer_mapping (
    id bigint auto_increment primary key,
    transaction_id varchar(40) not null,
    withdrawal_id bigint not null
) collate = utf8mb4_general_ci;
create index IDX_withdrawal_transfer_mapping_withdrawal_id on withdrawal_transfer_mapping (withdrawal_id);
#-------------------------v1.0.27---------------------------------------
alter table transfer_request
add completed_transfer varchar(255) null;
alter table transfer
add transfer_status varchar(255) null;
alter table transfer
add repayment_date datetime null;
alter table job_execute_time
add transfer_status_time datetime null;
alter table transfer
add error_code varchar(255) null;
alter table transfer
add error_msg varchar(255) null;
#-------------------------v1.1.0---------------------------------------
alter table sms_msg drop column receiver;
create table map_sms_receiver (
    id bigint auto_increment primary key,
    sms_id int not null,
    receiver varchar(50) not null,
    created datetime not null
);
#-------------------------v1.3.0---------------------------------------
update user
set investment_purpose = 'RAISING_FUTURE_MONEY'
where investment_purpose = 'RAISING_FUTURE_MANEY';
create table if not exists `column` (
    id bigint auto_increment primary key,
    title varchar(255) not null,
    content longtext not null,
    user_id bigint not null,
    thumbnail_url varchar(255) not null,
    is_posted tinyint(1) default 0 not null,
    created datetime not null,
    updated datetime null,
    deleted datetime null
);
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
);
create index column_file_column_id_index on column_file (column_id);
#-------------------------v1.4.0---------------------------------------
#-------------------------v1.5.0---------------------------------------
alter table `column`
add common_id varchar(40) not null;
#-------------------------v1.6.0---------------------------------------
#-------------------------v1.7.0---------------------------------------
alter table user change is_online_certification is_ekyc tinyint(1) null;
alter table sign_up_request change expire email_expire datetime null;
alter table sign_up_request
add url_expire datetime null;
drop table ekyc_request cascade;
create table if not exists ekyc_request_log (
    id bigint auto_increment primary key,
    email varchar(255) not null,
    request_id varchar(40) not null,
    ekyc_id varchar(40) not null,
    applicant_id bigint not null,
    status varchar(10) not null,
    created datetime not null
);
create index ekyc_request_log_request_id_index on ekyc_request_log (request_id);
alter table sign_up_request change expire email_expire datetime null;
alter table sign_up_request
add url_expire datetime null;
#-------------------------v1.8.0---------------------------------------
DROP TABLE jwt_skip_url cascade;
#-------------------------v1.9.0---------------------------------------
alter table project
add offering_type varchar(50) null;
alter table energysharesdb.project_loan_execute change loan_execute_status execute_status varchar(255) null;
alter table energysharesdb.project_loan_execute change loan_amount amount bigint null;
alter table energysharesdb.project_loan_execute change loan_execute_date execute_date datetime null;
alter table energysharesdb.project_loan_execute
    rename project_investment_execute;
create table event_message (
    event_category varchar(255) null,
    content blob null,
    uuid varchar(40) not null primary key,
    event_time datetime null,
    is_publish tinyint(1) null,
    created datetime null
);
alter table energysharesdb.share
add commitment_date datetime null;
alter table energysharesdb.dividend
add confirmed_date datetime null;
drop table mail_template cascade;
create table map_mail_receiver (
    id bigint auto_increment primary key,
    mail_id int not null,
    receiver varchar(50) not null,
    created datetime not null
);
-- mail 테이블의 receiver 컬럼 내용 map_mail_receiver 로 마이그레이션 쿼리
INSERT INTO map_mail_receiver (mail_id, receiver, created)
SELECT id,
    TRIM(
        BOTH ' '
        FROM TRIM(
                BOTH ','
                FROM SUBSTRING_INDEX(
                        SUBSTRING_INDEX(receiver, ',', n.digit + 1),
                        ',',
                        -1
                    )
            )
    ) as receiver_value,
    created
FROM mail
    JOIN (
        SELECT 0 as digit
        UNION ALL
        SELECT 1
        UNION ALL
        SELECT 2
    ) n ON LENGTH(receiver) - LENGTH(REPLACE(receiver, ',', '')) >= n.digit
WHERE TRIM(
        BOTH ' '
        FROM TRIM(
                BOTH ','
                FROM SUBSTRING_INDEX(
                        SUBSTRING_INDEX(receiver, ',', n.digit + 1),
                        ',',
                        -1
                    )
            )
    ) <> ''
ORDER BY id,
    n.digit;
alter table mail drop column receiver;
alter table marketing_code
add user_id bigint null;
alter table marketing_code drop column access_key;
#-------------------------v1.10.0---------------------------------------
DROP TABLE report_output_group;
DROP TABLE report_file;
DROP TABLE user_quarter_balance;
alter table energysharesdb.report_output drop column report_file_id;
alter table energysharesdb.report_output
add at_open datetime null;
#-------------------------v1.11.0---------------------------------------
#-------------------------v1.12.0---------------------------------------
alter table ekyc_request_log
add result varchar(10) null;
alter table ekyc_request_log
add error_message longtext null;
update ekyc_request_log
set result = 'MATCH'
where STATUS = 'SUCCESS'
    and request_id in (
        select request_id
        from user
        where audit_status = 'COMPLETED_AUDIT'
    );
update ekyc_request_log
set result = 'UNMATCH'
where STATUS = 'SUCCESS'
    and request_id in (
        select request_id
        from user
        where audit_status != 'COMPLETED_AUDIT'
    );
-- 콜레이션 통일
ALTER TABLE private_profile CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
-- 개인 회원 CHECK_ADDRESS_AUDIT -> HOLD_AUDIT
UPDATE user
SET audit_status = 'HOLD_AUDIT'
WHERE user_type = 'PRIVATE'
    AND audit_status = 'CHECK_ADDRESS_AUDIT';
-- 중복 파일 제거
DELETE ufa1
FROM user_file ufa1
    JOIN (
        SELECT user_id,
            file_type,
            MAX(id) AS max_id
        FROM user_file
        WHERE file_type IN ('IDENTIFICATION1', 'IDENTIFICATION2')
        GROUP BY user_id,
            file_type
    ) ufa2 ON ufa1.user_id = ufa2.user_id
    AND ufa1.file_type = ufa2.file_type
    AND ufa1.id < ufa2.max_id;
-- 파일 타입 수정
UPDATE user_file
SET file_type = 'IDENTIFICATION1_FRONT'
WHERE file_type = 'IDENTIFICATION1';
UPDATE user_file
SET file_type = 'IDENTIFICATION2_FRONT'
WHERE file_type = 'IDENTIFICATION2';
update user_changed_content
set content_type = 'USER_CHANGED_CONTENT_IDENTIFICATION1_FRONT'
where content_type = 'USER_CHANGED_CONTENT_IDENTIFICATION';
update user_changed_content
set content_type = 'USER_CHANGED_CONTENT_IDENTIFICATION2_FRONT'
where content_type = 'USER_CHANGED_CONTENT_IDENTIFICATION2';
-- 법인/개인 투자자 구분 두 개로 변경 관련 수정
UPDATE user
SET investor_type = 'PRIVATE_INVESTOR'
WHERE user_type = 'ADMIN';
UPDATE user
SET investor_type = 'PRIVATE_INVESTOR'
WHERE user_type = 'PRIVATE';
UPDATE user
SET investor_type = 'CORPORATE_INVESTOR'
WHERE user_type = 'CORPORATE';
UPDATE user
SET investor_type = SUBSTRING(investor_type, LENGTH('TEMPORARY_') + 1)
WHERE investor_type LIKE 'TEMPORARY_%';
-- 대출형 가상 계좌 삭제
DELETE FROM user_virtual_account
WHERE project_type = 'LOAN';
create table if not exists repayment_dividend_schedule (
    -- 기본정보
    id bigint auto_increment primary key,
    project_id bigint null,
    project_type varchar(255) null,
    -- 사업형(EQUITY), 대출형(LOAN). 현재는 사업형만 있다
    project_share_qty bigint null,
    -- 전체 구좌수
    project_share_amount bigint null,
    -- 전체 투자액
    -- 공통정보
    decide_date datetime null,
    -- 배당, 상환을 결정한 일시
    execute_date datetime null,
    -- 배당, 상환을 실시할 일시
    return_type varchar(255) null,
    -- DIVIDEND, REPAYMENT
    status varchar(255) not null,
    -- SCHEDULED, COMPLETED
    -- 상환정보
    repayment_amount bigint null,
    -- 상환액
    repayment_per_share bigint null,
    -- 구좌당 상환액
    previous_repayment_reserve bigint null,
    -- 이전 상환적립금
    repayment_reserve bigint null,
    -- 상환적립금
    -- 배당정보
    dividend_amount bigint null,
    -- 배당액
    net_dividend_after_broker_fee bigint null,
    -- 공제후 순배당액
    broker_fee bigint null,
    -- 공제액
    return_rate float null,
    -- 수익율
    dividend_per_share bigint null,
    -- 구좌당 배당액(공제후 금액. 세전)
    withholding_tax bigint null,
    -- 원천징수예치액
    investment_balance bigint null,
    -- 원금잔액
    -- 기본날짜정보
    created datetime null,
    updated datetime null,
    deleted datetime null
) collate = utf8mb4_general_ci;
#-------------------------v1.13.0---------------------------------------
-- consult_note 테이블 정리
alter table consult_note
modify user_id bigint not null;
alter table consult_note
modify owner_id bigint not null;
alter table consult_note change update_id modifier_id bigint not null;
alter table consult_note change creator owner varchar(50) not null;
alter table consult_note change request title varchar(255) not null;
alter table consult_note change memo content longtext not null;
drop index IDX_consult_note_owner_id on consult_note;
drop index IDX_consult_note_update_id on consult_note;
create index IDX_consult_note_user_id on consult_note (user_id);
#-------------------------v1.14.0---------------------------------------
-- japan bank 테이블 정리
drop table japan_bank cascade;
drop table japan_bank_branch cascade;
-- repayment_dividend_schedule 테이블에 법정장부를 위한 컬럼 추가
alter table repayment_dividend_schedule
add project_spv_id bigint null,
    add num_investors bigint null;
create table blog_post (
    id bigint auto_increment primary key,
    title varchar(255) not null,
    link varchar(255) not null,
    data text null,
    category varchar(255) null,
    source_url varchar(255) null,
    write_at datetime null,
    created datetime null,
    updated datetime null
);
-- user_asset_view 뷰 생성
CREATE VIEW user_asset_view AS WITH invest_operation_amount AS (
    SELECT owner_id,
        COALESCE(SUM(amount), 0) AS invest_operation_amount
    FROM share
    WHERE invest_status = 'OPERATION_INVEST'
    GROUP BY owner_id
),
invest_success_amount AS (
    SELECT owner_id,
        COALESCE(SUM(amount), 0) AS invest_success_amount
    FROM share
    WHERE invest_status = 'SUCCESS_INVEST'
    GROUP BY owner_id
),
invest_request_amount AS (
    SELECT s.owner_id,
        COALESCE(SUM(d.deposit_amount), 0) AS invest_request_amount
    FROM share s
        INNER JOIN deposit_request d ON s.id = d.share_id
    WHERE s.invest_status = 'REQUEST_INVEST'
    GROUP BY s.owner_id
),
withdrawal_request_amount AS (
    SELECT investor_id,
        COALESCE(SUM(amount), 0) AS withdrawal_request_amount
    FROM withdrawal
    WHERE status IN ('INCOMPLETE', 'WAIT', 'FAIL')
    GROUP BY investor_id
),
unpaid_amount AS (
    SELECT user_id,
        COALESCE(SUM(request_amount - deposit_amount), 0) AS unpaid_amount
    FROM deposit_request
    WHERE deposit_check_status = 'NO_DEPOSIT'
        AND expire_deposit_date >= CURRENT_TIMESTAMP
    GROUP BY user_id
)
SELECT u.id AS user_id,
    COALESCE(ua.balance, 0) AS deposit,
    COALESCE(ioa.invest_operation_amount, 0) AS invest_operation_amount,
    COALESCE(isa.invest_success_amount, 0) AS invest_success_amount,
    COALESCE(ira.invest_request_amount, 0) AS invest_request_amount,
    COALESCE(wra.withdrawal_request_amount, 0) AS withdrawal_request_amount,
    COALESCE(uaa.unpaid_amount, 0) AS unpaid_amount
FROM user u
    LEFT JOIN user_account ua ON u.id = ua.user_id
    LEFT JOIN invest_operation_amount ioa ON u.id = ioa.owner_id
    LEFT JOIN invest_success_amount isa ON u.id = isa.owner_id
    LEFT JOIN invest_request_amount ira ON u.id = ira.owner_id
    LEFT JOIN withdrawal_request_amount wra ON u.id = wra.investor_id
    LEFT JOIN unpaid_amount uaa ON u.id = uaa.user_id;
-- deposit_request_lack_history 의 잘못된 user_id 값 수정
UPDATE deposit_request_lack_history
    LEFT JOIN deposit_request ON deposit_request.id = deposit_request_lack_history.deposit_request_id
SET deposit_request_lack_history.user_id = deposit_request.user_id
WHERE deposit_request.id = deposit_request_lack_history.deposit_request_id;
#-------------------------v1.15.0---------------------------------------
#-------------------------v1.16.0---------------------------------------