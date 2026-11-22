DELETE FROM faqs WHERE sort_no >= 4;
DELETE FROM listings WHERE id BETWEEN 13 AND 30;
DELETE FROM destinations WHERE id IN (7, 8);
