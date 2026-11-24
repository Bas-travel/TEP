-- Platform-wide settings editable by admins (defaults live in code; a row exists only once an admin changes a value).
CREATE TABLE platform_settings (
  setting_key VARCHAR(60) NOT NULL PRIMARY KEY,
  setting_value VARCHAR(500) NOT NULL,
  updated_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
