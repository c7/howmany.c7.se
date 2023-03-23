#!/bin/env sh
/usr/bin/mlr --icsv --ojson sort -n index data/companies.csv | sed "s/\&amp;/\&/g" | jq . > data/companies.json
