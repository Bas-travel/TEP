ALTER TABLE bookings
  DROP CONSTRAINT ck_booking_guests,
  DROP CONSTRAINT ck_booking_dates,
  DROP COLUMN guests,
  DROP COLUMN end_date,
  DROP COLUMN start_date;
