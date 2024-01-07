-- 2023-12-18 commit: 21356038c43c84129bb2c2e5038cff172617a9e0
alter table account_records drop index tx_hash;
alter table account_records add column op_index int unsigned not null default 0 after `tx_index`;
alter table account_records add unique tx_hash_op_index (`tx_hash`,`op_index`);

-- 2024-01-01

ALTER TABLE `account_records` CHANGE `input` `input` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL;

-- 2023-01-05

CREATE TABLE IF NOT EXISTS `allowance` (
    `id` bigint unsigned NOT NULL,
    `tick` varchar(42) NOT NULL DEFAULT '',
    `tick_hash` varchar(66) NOT NULL DEFAULT '',
    `owner` varchar(42) NOT NULL DEFAULT '',
    `spender` varchar(42) NOT NULL DEFAULT '',
    `amt` varchar(128) NOT NULL DEFAULT '0',
    `position` varchar(42) NOT NULL DEFAULT '',
    `create_at` bigint unsigned NOT NULL DEFAULT '0',
    `update_at` bigint unsigned NOT NULL DEFAULT '0',
    `delete_at` bigint unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `owner_spender_tk` (`owner`, `spender`, `tick_hash`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

