-- +goose up

CREATE TABLE IF NOT EXISTS locale.localeitems_list (
  aggregate_id varchar(64) NOT NULL,
  updated_at timestamptz NOT NULL,
  updated_by varchar(64) NOT NULL,
  lang varchar(12) NOT NULL,
  content text NOT NULL,
  context varchar(64) NOT NULL,
  is_lang_reference boolean DEFAULT true,
  CONSTRAINT localeitems_list_pkey PRIMARY KEY (aggregate_id, lang, context)
);

-- +goose down
DROP TABLE locale.localeitems_list;
