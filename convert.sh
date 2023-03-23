#!/bin/env sh
/usr/bin/mlr --icsv --ojson sort -n index data/companies.csv | jq . > data/companies.json
