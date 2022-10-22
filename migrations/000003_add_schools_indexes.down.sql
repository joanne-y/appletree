-- Filename: migrations/000003_add_schools_indexes.down.sql
DROP INDEX If EXISTS schools_name_idx;
DROP INDEX If EXISTS schools_level_idx;
DROP INDEX If EXISTS schools_mode_idx;