-- 2023-12-18 commit: 21356038c43c84129bb2c2e5038cff172617a9e0
alter table account_records drop index tx_hash;
alter table account_records add column op_index int unsigned not null default 0 after `tx_index`;
alter table account_records add unique tx_hash_op_index (`tx_hash`,`op_index`);

-- 2024-01-01

ALTER TABLE `account_records` CHANGE `input` `input` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL;
