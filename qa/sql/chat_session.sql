create table `chat_session` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '聊天会话ID',
    `title` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '聊天会话标题',
    `user_id` bigint unsigned NOT NULL COMMENT '用户ID',

    `empty_slot` TINYINT NULL DEFAULT NULL COMMENT '空会话唯一占位(1=空会话,NULL=非空)',

    `message_count`     bigint unsigned NOT NULL DEFAULT 0 COMMENT '消息数量',
    `has_message`       tinyint NOT NULL DEFAULT 0 COMMENT '是否有消息 0没有 1有',
    `session_status`    tinyint NOT NULL DEFAULT 0 COMMENT '0空 1进行中 2结束',

    `is_pinned` TINYINT NOT NULL DEFAULT 0 COMMENT '是否置顶 0否 1是',
    `pinned_at` DATETIME NULL COMMENT '置顶时间',
    `last_message_at` datetime NOT NULL COMMENT '最后消息时间',

    `relation_status` TINYINT NOT NULL DEFAULT 0 COMMENT '0正常 1待同步',

    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间（系统时间）',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间（系统时间）',
    `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间戳(0=未删除，>0=删除时间)',

    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_empty` (`user_id`, `empty_slot`, `deleted_at`),
    KEY `idx_last_message_at` (`last_message_at`),
    KEY `idx_user_deleted` (`user_id`, `deleted_at`),
    KEY `idx_user_empty` (`user_id`, `session_status`, `deleted_at`),
    KEY idx_user_list (user_id, is_pinned, deleted_at, empty_slot,last_message_at DESC,id DESC)

)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天会话表';