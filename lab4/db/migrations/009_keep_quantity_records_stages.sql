begin;

-- Ensure quantity column exists to keep compatibility with handlers/templates
alter table if exists records_stages
  add column if not exists quantity integer not null default 1;

commit;
