-- +goose up

CREATE TABLE IF NOT EXISTS locale.user(
  subject_id varchar(128) NOT NULL,
  email varchar(128) NOT NULL,
  name varchar(128) NOT NULL,
  picture varchar(512),
  role int NOT NULL DEFAULT 2, -- translator role
  refresh_token varchar(512),
  CONSTRAINT user_subject_pkey PRIMARY KEY (subject_id )
);

-- +goose down
DROP TABLE locale.user;
