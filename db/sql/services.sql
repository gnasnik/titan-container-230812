CREATE TABLE IF NOT EXISTS services(
    id INT UNSIGNED AUTO_INCREMENT,
    image VARCHAR(128) NOT NULL,
    port INT DEFAULT 0,
    cpu FLOAT        DEFAULT 0,
    memory FLOAT        DEFAULT 0,
    storage FLOAT        DEFAULT 0,
    deployment_id VARCHAR(128) NOT NULL,
    created_at DATETIME     DEFAULT NULL,
    updated_at DATETIME     DEFAULT NULL,
    PRIMARY KEY (id)
    )ENGINE=InnoDB COMMENT='services';