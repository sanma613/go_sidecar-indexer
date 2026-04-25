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

go run . --mode search --query "AI AND LLM"
go run . --mode search --query "Prompt AND engineering"
go run . --mode search --query "pipelines AND automating AND Infracode"
go run . --mode search --query "LLM"
go run . --mode search --query "Architectures"
