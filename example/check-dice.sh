#!/usr/bin/env bash

num=$((RANDOM % 6 + 1))

if [ "$num" -gt 5 ]; then
  printf "%s" "$num"
  exit 2
elif [ "$num" -gt 3 ]; then
  printf "%s" "$num"
  exit 1
else
  printf "%s" "$num"
  exit 0
fi
