-- +goose up
-- \set psw_locale_user `echo $LOCALEMGMT_USER_PSW`

CREATE USER localemgmt_user
WITH
  LOGIN ENCRYPTED PASSWORD '${LOCALEMGMT_USER_PSW}';

CREATE SCHEMA locale AUTHORIZATION localemgmt_user;

-- +goose down
DROP SCHEMA locale;
DROP USER localemgmt_user;
