CREATE SCHEMA IF NOT EXISTS reposcan;

CREATE TYPE reposcan.scanning_status AS ENUM (
	'queued',
	'in_progress',
	'success',
	'failure');

CREATE TABLE reposcan.repositories (
    repository_id serial NOT NULL,
    repository_name varchar NOT NULL,
    repository_url varchar NOT NULL,
    is_active bit(1) NOT NULL DEFAULT '1'::"bit",
    created_by varchar NOT NULL DEFAULT 'SYSTEM'::character varying,
    created_at timestamp NOT NULL DEFAULT now(),
    modified_by varchar NOT NULL DEFAULT 'SYSTEM'::character varying,
    modified_at timestamp NOT NULL DEFAULT now(),
    deleted_by varchar,
    deleted_at timestamp,
    CONSTRAINT repositories_pkey PRIMARY KEY (repository_id)
);
CREATE INDEX repositories_repository_name_idx ON reposcan.repositories USING btree(repository_name);
CREATE INDEX repositories_is_active_idx ON reposcan.repositories USING btree(is_active);

CREATE TABLE reposcan.scannings (
    scanning_id serial NOT NULL,
    repository_id bigint NOT NULL,
    findings jsonb NOT NULL DEFAULT '{}'::jsonb,
    scanning_status reposcan.scanning_status NOT NULL DEFAULT 'queued'::reposcan.scanning_status,
    queued_at timestamp NOT NULL DEFAULT now(),
    scanning_at timestamp,
    finished_at timestamp,
    is_active bit(1) NOT NULL DEFAULT '1'::"bit",
    created_by varchar NOT NULL DEFAULT 'SYSTEM'::character varying,
    created_at timestamp NOT NULL DEFAULT now(),
    modified_by varchar NOT NULL DEFAULT 'SYSTEM'::character varying,
    modified_at timestamp NOT NULL DEFAULT now(),
    deleted_by varchar,
    deleted_at timestamp,
    CONSTRAINT scannings_pkey PRIMARY KEY (scanning_id),
    CONSTRAINT scannings_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES reposcan.repositories(repository_id)
);
CREATE INDEX scannings_repository_id_idx ON reposcan.scannings USING btree(repository_id);
CREATE INDEX scannings_scanning_status_idx ON reposcan.scannings USING btree(scanning_status);
CREATE INDEX scannings_is_active_idx ON reposcan.scannings USING btree(is_active);

