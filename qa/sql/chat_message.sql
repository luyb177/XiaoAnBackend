CREATE TABLE `chat_message` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息ID',

    `message_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '消息唯一ID（user 客户端生成, Assistant 服务端生成）',

    `session_id` BIGINT UNSIGNED NOT NULL COMMENT '会话ID',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT '发送者用户ID',

    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0生成中 1完成 2失败',

    `prompt_tokens` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '提示词令牌数（role为assistant时有效）',
    `completion_tokens` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成词令牌数（role为assistant时有效）',
    `total_tokens` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '总令牌数（role为assistant时有效）',

    `role` TINYINT NOT NULL DEFAULT 0 COMMENT '0用户 1AI 2系统',
    `message_type` TINYINT NOT NULL DEFAULT 0 COMMENT '0文本 1图片 2文件 3系统',

    `content` MEDIUMTEXT NOT NULL COMMENT '消息内容',

    `finish_reason` VARCHAR(32) NOT NULL DEFAULT '' COMMENT 'stop/length/tool',

    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间戳(0=未删除)',

    PRIMARY KEY (`id`),

    UNIQUE KEY `uk_session_client_msg` (`session_id`, `message_id`),

    KEY idx_user_status_session_deleted_id (`user_id`,`status`,`session_id`, `deleted_at`, `id`),
    KEY `idx_user_deleted` (`user_id`, `deleted_at`),
    KEY `idx_session_role_id` (`session_id`, `role`, `id`)

) ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_unicode_ci
COMMENT='聊天消息表';