-- Filename: migrations/000002_add_schools_check_constraint.down.sql

ALTER TABLE schools DROP CONSTRAINT IF EXISTS mode_length_check;