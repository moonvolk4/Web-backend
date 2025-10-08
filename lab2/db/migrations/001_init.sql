-- For visibility only: schema is already applied in the live DB.
-- stages, requests, request_stages, users

begin;

create table if not exists users (
  id           bigserial primary key,
  login        varchar(50)  not null unique,
  password     varchar(200) not null,
  is_moderator boolean      not null default false
);

create table if not exists stages (
  id           bigserial    primary key,
  title        varchar(100) not null,
  description  text         not null,
  is_deleted   boolean      not null default false,
  pressure     text         not null,
  risk_name    text         not null,
  risk_class   text         not null,
  code         text         not null,
  icon         text         not null,
  image_key    text         null
);

create index if not exists idx_stages_title  on stages(title);
create index if not exists idx_stages_active on stages(is_deleted);

create table if not exists records (
  id           bigserial primary key,
  status       text        not null check (status in ('черновик','удалён','сформирован','завершён','отклонён')),
  created_at   timestamptz not null default now(),
  formed_at    timestamptz null,
  finished_at  timestamptz null,
  creator_id   bigint      not null references users(id) on update cascade on delete restrict,
  moderator_id bigint      null     references users(id) on update cascade on delete restrict,
  total_items  integer     not null default 0
);

create index if not exists idx_requests_status  on requests(status);
create index if not exists idx_requests_creator on requests(creator_id);

create unique index if not exists idx_unique_draft_per_user
  on records(creator_id)
  where status = 'черновик';

create table if not exists records_stages (
  id               bigserial primary key,
  application_id   bigint  not null references records(id) on update cascade on delete restrict,
  order_id         bigint  not null references stages(id)   on update cascade on delete restrict,
  quantity         integer not null default 1 check (quantity > 0),
  position         integer not null default 0,
  is_primary       boolean not null default false,
  doctor_comment   text    null,
  selected_flags   jsonb   not null default '{}'::jsonb,
  doctor_id        bigint  null references users(id) on update cascade on delete restrict,
  constraint idx_app_order unique (application_id, order_id)
);

create index if not exists idx_req_stages_req   on records_stages(application_id);
create index if not exists idx_req_stages_stage on records_stages(order_id);

commit;
