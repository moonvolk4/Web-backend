-- For visibility only: trigger to maintain total_items on finish

begin;

create or replace function records_set_total_items_on_finish()
returns trigger as $$
begin
  if new.status = 'завершён' and (old.status is distinct from new.status) then
    update records r
    set total_items = coalesce((select sum(quantity) from records_stages rs where rs.application_id = r.id), 0)
    where r.id = new.id;
  end if;
  return new;
end;
$$ language plpgsql;

drop trigger if exists trg_rec_set_total_items on records;
create trigger trg_rec_set_total_items
after update of status on records
for each row
execute function records_set_total_items_on_finish();

commit;
