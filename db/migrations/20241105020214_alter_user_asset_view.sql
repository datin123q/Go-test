-- +goose Up
drop view user_asset_view;


create view user_asset_view as
select `u`.`id`                                         AS `user_id`,
       coalesce(`ua`.`balance`, 0)                      AS `deposit`,
       coalesce(`ioa`.`invest_operation_amount`, 0)     AS `invest_operation_amount`,
       coalesce(`isa`.`invest_success_amount`, 0)       AS `invest_success_amount`,
       coalesce(`ira`.`invest_request_amount`, 0)       AS `invest_request_amount`,
       coalesce(`wra`.`withdrawal_request_amount`, 0)   AS `withdrawal_request_amount`,
       coalesce(`uaa`.`unpaid_amount`, 0)               AS `unpaid_amount`,
       ((((coalesce(`ua`.`balance`, 0) + coalesce(`ioa`.`invest_operation_amount`, 0)) +
          coalesce(`isa`.`invest_success_amount`, 0)) + coalesce(`ira`.`invest_request_amount`, 0)) +
        coalesce(`wra`.`withdrawal_request_amount`, 0)) AS `total_asset`
from ((((((`user` `u` left join `user_account` `ua`
           on ((`u`.`id` = `ua`.`user_id`))) left join (select `share`.`owner_id`                 AS `owner_id`,
                                                               coalesce(sum(`share`.`amount`), 0) AS `invest_operation_amount`
                                                        from `share`
                                                        where (`share`.`invest_status` = 'OPERATION_INVEST')
                                                        group by `share`.`owner_id`) `ioa`
          on ((`u`.`id` = `ioa`.`owner_id`))) left join (select `share`.`owner_id`                 AS `owner_id`,
                                                                coalesce(sum(`share`.`amount`), 0) AS `invest_success_amount`
                                                         from `share`
                                                         where (`share`.`invest_status` = 'SUCCESS_INVEST')
                                                         group by `share`.`owner_id`) `isa`
         on ((`u`.`id` = `isa`.`owner_id`))) left join (select `s`.`owner_id`                         AS `owner_id`,
                                                               coalesce(sum(`d`.`deposit_amount`), 0) AS `invest_request_amount`
                                                        from (`share` `s` join `deposit_request` `d`
                                                              on ((`s`.`id` = `d`.`share_id`)))
                                                        where (`s`.`invest_status` = 'REQUEST_INVEST')
                                                        group by `s`.`owner_id`) `ira`
        on ((`u`.`id` = `ira`.`owner_id`))) left join (select `withdrawal`.`investor_id`              AS `investor_id`,
                                                              coalesce(sum(`withdrawal`.`amount`), 0) AS `withdrawal_request_amount`
                                                       from `withdrawal`
                                                       where (`withdrawal`.`status` in
                                                              ('INCOMPLETE', 'WAIT', 'FAIL'))
                                                       group by `withdrawal`.`investor_id`) `wra`
       on ((`u`.`id` = `wra`.`investor_id`))) left join (select `deposit_request`.`user_id` AS `user_id`,
                                                                coalesce(sum((
                                                                    `deposit_request`.`request_amount` -
                                                                    `deposit_request`.`deposit_amount`)),
                                                                         0)                                 AS `unpaid_amount`
                                                         from `deposit_request`
                                                         where ((`deposit_request`.`deposit_check_status` = 'NO_DEPOSIT') and
                                                                (`deposit_request`.`expire_deposit_date` >= now()))
                                                         group by `deposit_request`.`user_id`) `uaa`
      on ((`u`.`id` = `uaa`.`user_id`)));



