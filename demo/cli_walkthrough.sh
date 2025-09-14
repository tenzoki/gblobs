#!/bin/bash
set -e
echo "+ CLI Demo: Plain store, file, string, stats, inspect, delete"
STORE=tmp_cli_gblobs_store
rm -rf "$STORE"
mkdir -p "$STORE"
echo test > demo_file.txt
../cmd/gblobs/gblobs putfile --store "$STORE" demo_file.txt
../cmd/gblobs/gblobs putstring --store "$STORE" "demo string"
ID=$(../cmd/gblobs/gblobs putstring --store "$STORE" "demo string")
../cmd/gblobs/gblobs get --store "$STORE" "$ID"
../cmd/gblobs/gblobs exists --store "$STORE" "$ID"
../cmd/gblobs/gblobs stats --store "$STORE"
../cmd/gblobs/gblobs inspect --store "$STORE"
../cmd/gblobs/gblobs delete --store "$STORE" "$ID"
../cmd/gblobs/gblobs exists --store "$STORE" "$ID"
../cmd/gblobs/gblobs purge --store "$STORE"
../cmd/gblobs/gblobs stats --store "$STORE"
rm -rf "$STORE" demo_file.txt
echo "CLI demo finished."
