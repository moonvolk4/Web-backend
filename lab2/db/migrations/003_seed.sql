-- Optional seed data
begin;
insert into users (login, password, is_moderator)
values ('demo', '$2a$10$hashplaceholder', false)
on conflict (login) do nothing;
commit;
