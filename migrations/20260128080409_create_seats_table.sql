-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS seats (
  id SERIAL PRIMARY KEY,
  row VARCHAR(2) NOT NULL,
  seat_number INT NOT NULL,
  hall_id INT NOT NULL REFERENCES halls(id) ON DELETE CASCADE,
  version INT NOT NULL,

  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
);

CREATE INDEX seats_hall_id_version_idx ON seats (hall_id, version);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS seats;
-- +goose StatementEnd
