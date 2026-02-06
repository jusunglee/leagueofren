#!/usr/bin/env bash
cd "$(dirname "$0")"
bash scripts/seed_csv.sh scripts/kr_names.csv http://localhost:80
