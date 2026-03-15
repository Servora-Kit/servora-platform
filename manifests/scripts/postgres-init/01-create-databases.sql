-- Create separate databases for IAM (Ent) and OpenFGA.
-- Runs only on first init (empty data dir). One user (servora) can access both.
CREATE DATABASE iam;
CREATE DATABASE openfga;
