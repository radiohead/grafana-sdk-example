#!/bin/sh
set -e

dir=$1
if [[ -z $dir  ]]; then
  dir="./definitions"
fi

for file in $(find "${dir}" -name "*.json" -type f); do
  jq . "${file}" > "${file}.tmp"
  mv "${file}.tmp" "${file}"
done

