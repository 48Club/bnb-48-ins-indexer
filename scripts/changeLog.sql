-- 2023-12-18
alter table account_records drop index tx_hash;
alter table account_records add column op_index int not null default 0;
alter table account_records add unique tx_hash_op_index (`tx_hash`,`op_index`);