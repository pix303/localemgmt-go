-- +goose up

-- subject_id , email , name , role, contexts
CREATE TABLE IF NOT EXISTS locale.user(
  subject_id varchar(128) NOT NULL,
  email varchar(128) NOT NULL,
  name varchar(128) NOT NULL,
  role int NOT NULL DEFAULT 2, -- translator role
  contexts varchar(64)[],
    CONSTRAINT user_pkey PRIMARY KEY (subject_id )
);

CREATE TABLE IF NOT EXISTS locale.session(
  subject_id varchar(128) NOT NULL,
  refresh_token varchar(2048) NOT NULL,
  refresh_counter int NOT NULL DEFAULT 0,
  session_id varchar(128) NOT NULL,
  session_expire_at timestamptz,
  archived boolean,
    CONSTRAINT session_pkey PRIMARY KEY (subject_id )
);

-- +goose down
DROP TABLE locale.user;
DROP TABLE locale.session;
