-- PREPARE DATABASE

CREATE TABLE observations
(
    uid serial NOT NULL,
    observationtime timestamp,
    windspeed real,
    winddirection varchar (3),
    significantwaveheight real,
    dominantwaveperiod integer,
    averageperiod real,
    meanwavedirection varchar (3),
    airtemperature real,
    watertemperature real
);

ALTER TABLE observations
   ADD COLUMN rowcreated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT clock_timestamp();
