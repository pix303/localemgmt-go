-- +goose up
CREATE USER eventstore_user
WITH
  LOGIN ENCRYPTED PASSWORD '${STORE_USER_PSW}';

CREATE SCHEMA store AUTHORIZATION eventstore_user;

CREATE TABLE IF NOT EXISTS store.events (
  id bigserial not null,
  aggregateId varchar(64) not null,
  aggregateName varchar(64) not null,
  createdAt timestamptz not null,
  createdBy varchar(64),
  eventType varchar(64),
  payloadData text,
  payloadDataType varchar(64),
  CONSTRAINT "eventstore_pkey" PRIMARY KEY (id),
  CONSTRAINT "eventstore_ukey" UNIQUE (id, aggregateId, createdAt)
);

CREATE INDEX eventstore_index ON store.events (aggregateId);

CREATE INDEX eventstore_type_index ON store.events (eventType);

-- +goose down
DROP INDEX IF EXISTS store.eventstore_index;
DROP INDEX IF EXISTS store.eventstore_type_index;
DROP TABLE store.events;
DROP SCHEMA store;
DROP USER eventstore_user;
