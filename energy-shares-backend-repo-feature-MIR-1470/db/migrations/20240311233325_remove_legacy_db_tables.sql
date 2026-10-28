-- +goose Up
DROP TABLE IF EXISTS repayment_settlement_mapping;
DROP TABLE IF EXISTS repayment_schedule;
DROP TABLE IF EXISTS settlement_schedule;
DROP TABLE IF EXISTS settlement;
-- +goose Down
CREATE TABLE `repayment_schedule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_id` bigint DEFAULT NULL,
  `repayment_principal` bigint DEFAULT NULL,
  `interest_amount` bigint DEFAULT NULL,
  `broker_fee` bigint DEFAULT NULL,
  `revenue` bigint DEFAULT NULL,
  `repayment_date` datetime DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  `deleted` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `IDX_repayment_schedule_project_id` (`project_id`)
) ENGINE = InnoDB AUTO_INCREMENT = 56 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci;
CREATE TABLE `repayment_settlement_mapping` (
  `repayment_schedule_id` bigint DEFAULT NULL,
  `settlement_schedule_id` bigint NOT NULL,
  `created` datetime DEFAULT NULL
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci;
CREATE TABLE `settlement` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_id` bigint DEFAULT NULL,
  `project_type` varchar(255) DEFAULT NULL,
  `schedule_id` bigint NOT NULL,
  `round` int DEFAULT NULL,
  `settlement_type` varchar(255) DEFAULT NULL,
  `status` varchar(255) NOT NULL,
  `share_dividend_amount` float DEFAULT NULL,
  `share_repayment_amount` float DEFAULT NULL,
  `project_share_qty` bigint DEFAULT NULL,
  `dividend_amount` bigint DEFAULT NULL,
  `interest_amount` bigint DEFAULT NULL,
  `repayment_amount` bigint DEFAULT NULL,
  `broker_fee` bigint DEFAULT NULL,
  `with_holding_tax` bigint DEFAULT NULL,
  `etc_profit` bigint DEFAULT NULL,
  `repayment_reserve` bigint DEFAULT NULL,
  `before_repayment_reserve` bigint DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  `deleted` datetime DEFAULT NULL,
  `project_share_amount` bigint DEFAULT NULL,
  `before_etc_profit` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `UQE_settlement_schedule_id` (`schedule_id`)
) ENGINE = InnoDB AUTO_INCREMENT = 44 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci;
CREATE TABLE `settlement_schedule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_id` bigint DEFAULT NULL,
  `start_date` datetime DEFAULT NULL,
  `end_date` datetime DEFAULT NULL,
  `dividend_date` datetime DEFAULT NULL,
  `is_confirm` tinyint(1) DEFAULT NULL,
  `execution_status` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE = InnoDB AUTO_INCREMENT = 44 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_general_ci;