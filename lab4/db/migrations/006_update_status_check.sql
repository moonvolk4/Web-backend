begin;

-- Relax and define allowed statuses to align with application logic
-- Drop old check constraint if exists (name may vary; using records_status_check per error)
alter table if exists records
  drop constraint if exists records_status_check;

alter table if exists records
  add constraint records_status_check
  check (status in ('черновик','сформирована','завершена','отклонена','удалён'));

commit;
