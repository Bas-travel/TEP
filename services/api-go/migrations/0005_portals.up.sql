-- Partner / Admin / Staff portals (docs/requirements/mvp-scope.md, docs/architecture/security-model.md).

-- A partner is a business account owned by one user (users.role = 'partner' once approved).
CREATE TABLE partners (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED NOT NULL,
  business_name VARCHAR(160) NOT NULL,
  contact_phone VARCHAR(32) NULL,
  description VARCHAR(500) NULL,
  status ENUM('pending','approved','rejected','suspended') NOT NULL DEFAULT 'pending',
  commission_rate DECIMAL(5,2) NOT NULL DEFAULT 12.00,
  decision_note VARCHAR(255) NULL,
  decided_by BIGINT UNSIGNED NULL,
  decided_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_partner_user (user_id),
  KEY ix_partner_status (status, created_at),
  CONSTRAINT ck_partner_commission CHECK (commission_rate BETWEEN 0 AND 50),
  CONSTRAINT fk_partner_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Products/services are owned by a partner (NULL = platform-owned, e.g. seed data).
ALTER TABLE listings
  ADD COLUMN partner_id BIGINT UNSIGNED NULL AFTER destination_id,
  ADD COLUMN created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ADD KEY ix_listing_partner (partner_id, status),
  ADD CONSTRAINT fk_listing_partner FOREIGN KEY (partner_id) REFERENCES partners(id);

ALTER TABLE bookings ADD COLUMN completed_at DATETIME NULL AFTER cancelled_at;

-- Support desk: ticket lifecycle and a reply thread (staff <-> traveler).
ALTER TABLE support_tickets
  ADD COLUMN status ENUM('open','answered','resolved') NOT NULL DEFAULT 'open',
  ADD COLUMN first_replied_at DATETIME NULL,
  ADD COLUMN resolved_at DATETIME NULL,
  ADD KEY ix_ticket_status (status, created_at);

CREATE TABLE support_ticket_replies (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ticket_id BIGINT UNSIGNED NOT NULL,
  author_id BIGINT UNSIGNED NOT NULL,
  author_role ENUM('traveler','staff','admin') NOT NULL,
  message VARCHAR(2000) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY ix_reply_ticket (ticket_id, id),
  CONSTRAINT fk_reply_ticket FOREIGN KEY (ticket_id) REFERENCES support_tickets(id) ON DELETE CASCADE,
  CONSTRAINT fk_reply_author FOREIGN KEY (author_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Staff verification of a booking (status / documents). Several checks per booking are allowed; the latest one counts.
CREATE TABLE booking_checks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  booking_id BIGINT UNSIGNED NOT NULL,
  staff_id BIGINT UNSIGNED NOT NULL,
  result ENUM('verified','flagged') NOT NULL,
  note VARCHAR(500) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY ix_check_booking (booking_id, id),
  KEY ix_check_staff (staff_id, created_at),
  CONSTRAINT fk_check_booking FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE,
  CONSTRAINT fk_check_staff FOREIGN KEY (staff_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Every mutation done through /admin/v1, /partner/v1 and /staff/v1 is recorded (audit_logs exists since 0001).
ALTER TABLE audit_logs
  ADD KEY ix_audit_action (action, created_at),
  ADD KEY ix_audit_created (created_at);
