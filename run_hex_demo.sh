#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

echo "== Limpieza de segmentos previos =="
rm -f data/dict-*.sst data/post-*.sst data/dict-*.sidx || true

echo "== Ingesta unica desde carpeta (auto-detect por extension) =="
echo "== Segmentacion forzada por memoria (threshold bajo) =="
go run . \
	--input-dir sample_inputs \
	--mode ingest \
	--mem-threshold-bytes 8192 \
	--mem-hard-limit-bytes 16384

echo "== Segmentos creados =="
ls -1 data/dict-*.sst data/post-*.sst

echo "== Query newest->oldest (AND) =="
go run . --mode search --query "introduccion and sidecar"

echo "== Query simple =="
go run . --mode search --query "hexagonal"
