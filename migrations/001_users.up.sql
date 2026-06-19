CREATE TABLE IF NOT EXISTS users (
    id                       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    email                    VARCHAR(255)    NOT NULL,
    password_hash            VARCHAR(255)    NOT NULL,
    name                     VARCHAR(100)    NOT NULL,
    refresh_token            VARCHAR(512)        NULL,
    created_at               DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at               DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY  uq_users_email         (email),
    INDEX       idx_users_refresh_token (refresh_token)
);
