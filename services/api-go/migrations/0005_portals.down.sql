ALTER TABLE audit_logs DROP KEY ix_audit_action, DROP KEY ix_audit_created;
DROP TABLE booking_checks;
DROP TABLE support_ticket_replies;
ALTER TABLE support_tickets
  DROP KEY ix_ticket_status,
  DROP COLUMN resolved_at,
  DROP COLUMN first_replied_at,
  DROP COLUMN status;
ALTER TABLE bookings DROP COLUMN completed_at;
ALTER TABLE listings DROP FOREIGN KEY fk_listing_partner;
ALTER TABLE listings DROP KEY ix_listing_partner, DROP COLUMN created_at, DROP COLUMN partner_id;
DROP TABLE partners;
