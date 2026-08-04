-- Run with psql against the shared Supabase project before applying the real migration.
-- The migration executes inside a transaction and is rolled back unconditionally.
\set ON_ERROR_STOP on
begin;
\ir ../migrations/0001_nodescope_foundation.sql
rollback;
