-- Execute as nodescope_migrate_login to prove it can apply only NodeScope DDL.
\set ON_ERROR_STOP on
begin;
\ir ../migrations/0003_schema_migration_history.sql
rollback;
