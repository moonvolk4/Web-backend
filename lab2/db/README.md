These SQL files are provided for lab deliverables. The running app already uses a live database; these are for reference only.

- migrations/001_init.sql — tables (users, stages, requests, request_stages) and indexes
- migrations/002_trigger.sql — trigger to recalc total_items on status change to 'завершён'
- migrations/003_seed.sql — minimal seed (demo user)

Apply order: 001, 002, 003.
