CREATE TABLE units (
  id            UUID              NOT NULL,
  class_id      UUID              NOT NULL,
  title         TEXT              NOT NULL,
  display_order INTEGER DEFAULT 0 NOT NULL
);
ALTER TABLE ONLY units
  ADD CONSTRAINT units_pkey PRIMARY KEY (id);
CREATE INDEX units_class_id_idx
  ON units USING BTREE (class_id);