-- +goose up

CREATE TABLE IF NOT EXISTS locale.localeitems (
  aggregateId varchar(64) not null,
  updatedAt timestamptz not null,
  data text,
  CONSTRAINT "localeitems_pkey" PRIMARY KEY (aggregateId)
);

-- +goose down
DROP TABLE locale.localeitems;
