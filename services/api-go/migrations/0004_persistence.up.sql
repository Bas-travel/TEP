-- Sprint 2: everything the API persists in MySQL (API_MODE=db). See docs/database/erd.md.

-- Public member handle used by the community API (e.g. m_nok, m_12); bio shown on member profiles.
ALTER TABLE users ADD COLUMN handle VARCHAR(40) NULL AFTER email;
UPDATE users SET handle = CONCAT('m_', id) WHERE handle IS NULL;
ALTER TABLE users MODIFY handle VARCHAR(40) NOT NULL, ADD UNIQUE KEY uk_users_handle (handle);
ALTER TABLE user_profiles ADD COLUMN bio VARCHAR(200) NULL AFTER level;

-- Booking type (hotel / restaurant / flight / attraction / trip) is shown in lists and detail.
ALTER TABLE bookings ADD COLUMN item_type VARCHAR(20) NOT NULL DEFAULT 'trip' AFTER status;

-- Idempotent replays must return the exact stored bytes; the JSON type would normalise them.
ALTER TABLE idempotency_keys MODIFY response_body MEDIUMTEXT NOT NULL;

ALTER TABLE community_groups ADD COLUMN rules JSON NULL AFTER gradient_css;
ALTER TABLE community_groups ADD UNIQUE KEY uk_group_name (name);
ALTER TABLE community_comments MODIFY body VARCHAR(1000) NOT NULL;

-- Per-user "hide this post from my feed".
CREATE TABLE post_hides (
  post_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (user_id, post_id),
  CONSTRAINT fk_ph_post FOREIGN KEY (post_id) REFERENCES community_posts(id) ON DELETE CASCADE,
  CONSTRAINT fk_ph_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Feed query: newest first, optionally per destination.
ALTER TABLE community_posts ADD KEY ix_post_feed (created_at, id), ADD KEY ix_post_dest (destination, created_at);
ALTER TABLE community_comments ADD KEY ix_comment_post (post_id, id), DROP KEY fk_cc_post;
