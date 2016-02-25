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

-- Extract CSV
\copy (SELECT observationtime, significantwaveheight, dominantwaveperiod, averageperiod, airtemperature, watertemperature FROM observations) TO data.csv CSV DELIMITER ',';

-- Select Max wave height ever
SELECT * FROM observations ORDER BY significantwaveheight DESC LIMIT 1;

-- Select Max wave height in given year
SELECT * FROM observations WHERE EXTRACT(year FROM "rowcreated") = 2016 ORDER BY significantwaveheight DESC LIMIT 1;
