-- +goose up

CREATE TABLE IF NOT EXISTS locale.localeitem_detail (
  aggregateId varchar(64) not null,
  updatedAt timestamptz not null,
  data text,
  CONSTRAINT "localeitems_pkey" PRIMARY KEY (aggregateId)
);

-- +goose down
DROP TABLE locale.localeitem_detail;
