CREATE TABLE files (
    id serial not null primary key,
    userid integer not null,
    filename text not null, -- with file path and extension
    uuid uuid not null default gen_random_uuid(),
    public bool default false,
    uploaded_at timestamp not null default now()
);