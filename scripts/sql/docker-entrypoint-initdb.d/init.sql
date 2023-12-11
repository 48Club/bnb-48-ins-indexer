CREATE DATABASE IF NOT EXISTS `fans_index` CHARACTER SET = utf8mb4;

USE `fans_index`;

CREATE TABLE IF NOT EXISTS `account` (
    `id` bigint unsigned NOT NULL,
    `address` varchar(128) NOT NULL DEFAULT '',
    `balance` bigint NOT NULL DEFAULT '0',
    `create_at` bigint unsigned NOT NULL DEFAULT '0',
    `update_at` bigint unsigned NOT NULL DEFAULT '0',
    `delete_at` bigint unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `address` (`address`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `account_records` (
    `id` bigint unsigned NOT NULL,
    `block` bigint NOT NULL DEFAULT '0',
    `tx_hash` varchar(128) NOT NULL DEFAULT '',
    `from` varchar(128) NOT NULL DEFAULT '',
    `to` varchar(128) NOT NULL DEFAULT '',
    `input` varchar(1024) NOT NULL DEFAULT '',
    `type` tinyint NOT NULL,
    `create_at` bigint unsigned NOT NULL DEFAULT '0',
    `update_at` bigint unsigned NOT NULL DEFAULT '0',
    `delete_at` bigint unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    UNIQUE KEY `tx_hash` (`tx_hash`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;

CREATE TABLE IF NOT EXISTS `account_hash` (
    `id` bigint unsigned NOT NULL,
    `account_id` bigint unsigned NOT NULL,
    `mint_hash` varchar(128) NOT NULL DEFAULT '',
    `state` tinyint NOT NULL,
    `create_at` bigint unsigned NOT NULL DEFAULT '0',
    `update_at` bigint unsigned NOT NULL DEFAULT '0',
    `delete_at` bigint unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
