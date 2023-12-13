CREATE DATABASE IF NOT EXISTS `fans_index` CHARACTER SET = utf8mb4;

USE `fans_index`;

CREATE TABLE IF NOT EXISTS `account` (
    `id` bigint unsigned NOT NULL,
    `address` varchar(42) NOT NULL DEFAULT '',
    `create_at` bigint unsigned NOT NULL DEFAULT '0',
    `update_at` bigint unsigned NOT NULL DEFAULT '0',
    `delete_at` bigint unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `address` (`address`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `account_wallet` (
    `id` bigint unsigned NOT NULL,
    `account_id` bigint unsigned NOT NULL,
    `tick` varchar(42) NOT NULL DEFAULT '',
    `tick_hash` varchar(66) NOT NULL DEFAULT '',
    `balance` varchar(128) NOT NULL DEFAULT '',
    `create_at` bigint unsigned NOT NULL DEFAULT '0',
    `update_at` bigint unsigned NOT NULL DEFAULT '0',
    `delete_at` bigint unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `account_id_tick_hash` (`account_id`,`tick_hash`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `inscription` (
    `id` bigint unsigned NOT NULL,
    `tick` varchar(42) NOT NULL DEFAULT '',
    `tick_hash` varchar(66) NOT NULL DEFAULT '',
    `tx_index` int unsigned NOT NULL,
    `block` bigint NOT NULL DEFAULT '0',
    `block_at` bigint unsigned NOT NULL DEFAULT '0',
    `decimal` tinyint unsigned NOT NULL,
    `max` varchar(128) NOT NULL DEFAULT '',
    `lim` varchar(128) NOT NULL DEFAULT '',
    `miners` varchar(2048) NOT NULL DEFAULT '',
    `minted` varchar(128) NOT NULL DEFAULT '',
    `create_at` bigint unsigned NOT NULL DEFAULT '0',
    `update_at` bigint unsigned NOT NULL DEFAULT '0',
    `delete_at` bigint unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `tick_hash` (`tick_hash`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `account_records` (
    `id` bigint unsigned NOT NULL,
    `block` bigint NOT NULL DEFAULT '0',
    `block_at` bigint unsigned NOT NULL DEFAULT '0',
    `tx_hash` varchar(66) NOT NULL DEFAULT '',
    `tx_index` int unsigned NOT NULL,
    `from` varchar(42) NOT NULL DEFAULT '',
    `to` varchar(42) NOT NULL DEFAULT '',
    `input` varchar(1024) NOT NULL DEFAULT '',
    `type` tinyint unsigned NOT NULL,
    `create_at` bigint unsigned NOT NULL DEFAULT '0',
    `update_at` bigint unsigned NOT NULL DEFAULT '0',
    `delete_at` bigint unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `tx_hash` (`tx_hash`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

INSERT INTO inscription (id,tick,tick_hash,tx_index,block,block_at,`decimal`,`max`,lim,miners,minted,create_at,update_at,delete_at)values("1","fans","0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2","237","34175786","1702042086","1","1000000","1","0x72b61c6014342d914470eC7aC2975bE345796c2b","1000000","1702042086000","0","0");
