-- F-1: persist stay/visit dates and guest count on a booking.
ALTER TABLE bookings
  ADD COLUMN start_date DATE NULL AFTER subtitle,
  ADD COLUMN end_date DATE NULL AFTER start_date,
  ADD COLUMN guests SMALLINT UNSIGNED NOT NULL DEFAULT 1 AFTER end_date,
  ADD CONSTRAINT ck_booking_dates CHECK (end_date IS NULL OR (start_date IS NOT NULL AND end_date >= start_date)),
  ADD CONSTRAINT ck_booking_guests CHECK (guests BETWEEN 1 AND 20);
