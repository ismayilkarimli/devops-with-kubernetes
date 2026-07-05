#!/usr/bin/env sh

random_string="$(openssl rand -hex 16)"

while true; do
  printf '%s\n' "$random_string"
  sleep 5
done
