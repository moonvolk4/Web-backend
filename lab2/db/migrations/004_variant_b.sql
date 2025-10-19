begin;

-- Stages: add 4 pressure fields (keep existing fields for now)
alter table if exists stages
  add column if not exists sys_from integer not null default 0,
  add column if not exists sys_to   integer not null default 0,
  add column if not exists dia_from integer not null default 0,
  add column if not exists dia_to   integer not null default 0;

-- Records: drop total_items if exists, add result fields and minimal flags
alter table if exists records
  drop column if exists total_items,
  add column if not exists result_stage_code text null,
  add column if not exists result_map numeric(5,2) null,
  add column if not exists flag_lv_hypertrophy boolean not null default false,
  add column if not exists flag_renal_damage boolean not null default false,
  add column if not exists flag_arterial_stiffness boolean not null default false;

-- Ensure unique draft per user (partial unique index)
drop index if exists idx_unique_draft_per_user;
create unique index idx_unique_draft_per_user on records(creator_id) where status = 'черновик';

-- Records_stages: keep only minimal fields
alter table if exists records_stages
  drop column if exists position,
  drop column if exists is_primary,
  drop column if exists selected_flags,
  drop column if exists doctor_id,
  add column if not exists quantity integer not null default 1,
  add column if not exists doctor_comment text null;

commit;
