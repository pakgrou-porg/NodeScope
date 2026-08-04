-- Validate the additive transactional-operation migration without persisting it.
\set ON_ERROR_STOP on
begin;
\ir ../migrations/0002_transactional_operations.sql
rollback;
