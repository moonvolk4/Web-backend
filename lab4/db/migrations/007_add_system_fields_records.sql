begin;

alter table if exists records
  add column if not exists formed_at timestamp null,
  add column if not exists finished_at timestamp null,
  add column if not exists moderator_id integer null;

commit;
