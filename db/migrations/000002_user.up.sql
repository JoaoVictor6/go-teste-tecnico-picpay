CREATE TYPE doc_type_enum AS ENUM (
    'CPF',
    'CNPJ'
);

CREATE TABLE client (
  id SERIAL PRIMARY KEY,
  name VARCHAR(150) NOT NULL,
  doc_type doc_type_enum NOT NULL,
  doc_value VARCHAR(14) UNIQUE NOT NULL,
  password VARCHAR(200) NOT NULL,
  wallet INT NOT NULL DEFAULT 0
);

