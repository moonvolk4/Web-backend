begin;

alter table if exists records
  add column if not exists comment text null;

commit;
