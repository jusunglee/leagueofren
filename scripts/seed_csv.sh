#!/usr/bin/env bash
# Bulk-seed translations from a 2-column CSV (USERNAME#TAG,REGION).
# Each row is POSTed to the web server's translation endpoint, which
# handles Riot validation + LLM translation.
#
# Usage:
#   bash scripts/seed_csv.sh translations.csv http://localhost:3000
#
# CSV format (no header row):
#   不知火舞#KR1,KR
#   小鱼人#NA1,NA

set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <csv-file> <server-url>"
  echo "  e.g. $0 translations.csv http://localhost:3000"
  exit 1
fi

CSV_FILE="$1"
SERVER_URL="${2%/}"  # strip trailing slash
ENDPOINT="$SERVER_URL/api/v1/translations"
DELAY=2  # seconds between requests to respect rate limiting

if [[ ! -f "$CSV_FILE" ]]; then
  echo "Error: file not found: $CSV_FILE"
  exit 1
fi

total=$(wc -l < "$CSV_FILE" | tr -d ' ')
current=0
success=0
fail=0

echo "Seeding $total entries from $CSV_FILE -> $ENDPOINT"
echo "Delay between requests: ${DELAY}s"
echo ""

while IFS=',' read -r username region || [[ -n "$username" ]]; do
  # Skip empty lines
  [[ -z "$username" ]] && continue

  # Trim whitespace
  username=$(echo "$username" | xargs)
  region=$(echo "$region" | xargs)

  current=$((current + 1))

  body=$(printf '{"username":"%s","region":"%s"}' "$username" "$region")

  http_code=$(curl -s -o /tmp/seed_csv_response.json -w '%{http_code}' \
    -X POST "$ENDPOINT" \
    -H 'Content-Type: application/json' \
    -d "$body")

  if [[ "$http_code" == "201" ]]; then
    translation=$(jq -r '.translation // "?"' /tmp/seed_csv_response.json 2>/dev/null || echo "?")
    echo "  [$current/$total] OK  $username ($region) -> $translation"
    success=$((success + 1))
  else
    error=$(jq -r '.error // "unknown"' /tmp/seed_csv_response.json 2>/dev/null || echo "HTTP $http_code")
    echo "  [$current/$total] FAIL $username ($region) — $error"
    fail=$((fail + 1))
  fi

  # Don't sleep after the last entry
  if [[ $current -lt $total ]]; then
    sleep "$DELAY"
  fi
done < "$CSV_FILE"

rm -f /tmp/seed_csv_response.json

echo ""
echo "Done! $success succeeded, $fail failed out of $total."
