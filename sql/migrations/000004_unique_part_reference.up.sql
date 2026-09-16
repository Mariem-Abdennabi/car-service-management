-- A part number that identifies two different parts identifies neither.
CREATE UNIQUE INDEX parts_reference_key ON parts (reference);
