-- Auto-create settings + session rows for every new user
create or replace function public.handle_new_user()
returns trigger language plpgsql security definer
set search_path = public
as $$
begin
  insert into public.settings (user_id) values (new.id);
  insert into public.session (user_id) values (new.id);
  return new;
end;
$$;

-- Drop existing trigger if re-running this migration
drop trigger if exists on_auth_user_created on auth.users;

create trigger on_auth_user_created
  after insert on auth.users
  for each row execute function public.handle_new_user();
