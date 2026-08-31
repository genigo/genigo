-- genigo test fixture: exercises every type/introspection feature on mysql
CREATE TABLE users (
  id BIGINT NOT NULL AUTO_INCREMENT,
  name VARCHAR(120) NOT NULL,
  email VARCHAR(190) NOT NULL,
  nickname VARCHAR(120) NULL,
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  status ENUM('active','pending','blocked') NOT NULL DEFAULT 'active',
  balance DECIMAL(12,2) NOT NULL DEFAULT 0,
  meta JSON NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB;

CREATE TABLE posts (
  id BIGINT NOT NULL AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  title VARCHAR(200) NOT NULL,
  slug VARCHAR(220) NOT NULL,
  body TEXT NULL,
  rating DECIMAL(4,2) NOT NULL DEFAULT 0,
  views INT NOT NULL DEFAULT 0,
  published_at DATETIME NULL,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uq_posts_slug (slug),
  CONSTRAINT fk_posts_user FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE=InnoDB;
