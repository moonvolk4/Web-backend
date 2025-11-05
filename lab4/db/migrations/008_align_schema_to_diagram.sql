begin;

-- Users table (for creator/moderator logins)
create table if not exists users (
  id bigserial primary key,
  login varchar(50) unique not null,
  password varchar(200) not null,
  is_moderator boolean not null default false
);

-- Records: add created_at, ensure FK to users, rename result_stage_code -> result_stage
alter table if exists records
  add column if not exists created_at timestamp not null default now(),
  add column if not exists result_stage text;

do $$ begin
  if exists (select 1 from information_schema.columns where table_name='records' and column_name='result_stage_code') then
    execute 'alter table records rename column result_stage_code to result_stage';
  end if;
end $$;

alter table if exists records
  add constraint if not exists records_creator_id_fkey foreign key (creator_id) references users(id);

alter table if exists records
  alter column moderator_id drop not null,
  add constraint if not exists records_moderator_id_fkey foreign key (moderator_id) references users(id);

-- Stages: soft delete flag
alter table if exists stages
  add column if not exists is_deleted boolean not null default false;

-- records_stages: ensure doctor_comment exists; drop quantity if present
alter table if exists records_stages
  add column if not exists doctor_comment text;

do $$ begin
  if exists (select 1 from information_schema.columns where table_name='records_stages' and column_name='quantity') then
    alter table records_stages drop column quantity;
  end if;
end $$;

commit;
