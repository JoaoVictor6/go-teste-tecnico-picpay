#!/bin/bash

 migrate -database "pgx5://username:password@localhost:5432/db" -path db/migrations $@
