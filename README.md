# Go Sidecar Indexer

Go Sidecar Indexer is a lightweight full-text indexing engine for CSV, JSON, and PDF documents. It builds immutable dictionary and postings segments on disk (SSTables), then runs term-based search queries across one or many segments. The project uses a memtable-based ingestion flow with sparse checkpoints for fast lookup.

## Features

- Ingest documents from a single file or a directory (`csv`, `json`, `pdf`, or `auto` detection).
- Build timestamped on-disk segments (`dict-*.sst`, `post-*.sst`) with sparse index checkpoints (`dict-*.sidx`).
- Search across the latest available segments in a directory or against an explicit dict/post pair.
- Support AND queries with normalized lowercase terms.

## Requirements

- Go `1.25.5` (from `go.mod`)

## Quick Start

### 1) Run ingest mode

Ingest one file:

```bash
go run . --mode ingest --input sample_inputs/hex_demo/lote1.json
```

Ingest all supported files in a directory:

```bash
go run . --mode ingest --input-dir sample_inputs/hex_demo
```

### 2) Run search mode

Search across discovered segments in `data/`:

```bash
go run . --mode search --query "introduccion and sidecar"
```

Search with an explicit segment pair:

```bash
go run . --mode search \
  --query "hexagonal" \
  --dict data/dict-1774374713257726675.sst \
  --post data/post-1774374713257726675.sst
```

### 3) Run ingest + search in one command

```bash
go run . --mode ingest-search --input-dir sample_inputs/hex_demo --query "sidecar and index"
```

## Build a Binary

```bash
go build -o sidecar-indexer .
./sidecar-indexer --mode ingest --input-dir sample_inputs/hex_demo
```

## Quick Demo (One Command)

Run the bundled demo script to execute end-to-end flow:

1. Clean previous generated segments in `data/`.
2. Ingest mixed sample files from `sample_inputs/hex_demo`.
3. Run search queries (`AND` and single-term examples).

```bash
bash run_hex_demo.sh
```

## CLI Flags

| Flag                     | Default            | Description                                                      |
| ------------------------ | ------------------ | ---------------------------------------------------------------- |
| `--mode`                 | `ingest`           | Execution mode: `ingest`, `search`, `ingest-search`              |
| `--format`               | `auto`             | Input parser format: `auto`, `csv`, `json`, `pdf`                |
| `--input`                | `""`               | Input file path                                                  |
| `--input-dir`            | `""`               | Input directory path (scan `csv/json/pdf`)                       |
| `--dict`                 | `""`               | Dictionary SST output path (must be used together with `--post`) |
| `--post`                 | `""`               | Postings SST output path (must be used together with `--dict`)   |
| `--segment-dir`          | `data`             | Segment directory for timestamped search/ingest outputs          |
| `--query`                | `""`               | Query string for search modes                                    |
| `--checkpoint-stride`    | `64`               | Sparse checkpoint interval                                       |
| `--mem-threshold-bytes`  | `54 * 1024 * 1024` | Memtable flush threshold                                         |
| `--mem-hard-limit-bytes` | `70 * 1024 * 1024` | Memtable hard limit (must be greater than threshold)             |

## Query Behavior

- Terms are normalized to lowercase.
- AND is supported using the string `and` (case-insensitive input, normalized internally).
- Example: `"go AND search and engine"`.

## Generated Files

During ingestion, the engine writes segment files (by default in `data/`):

- `dict-<timestamp>.sst` (dictionary entries)
- `post-<timestamp>.sst` (posting lists)
- `dict-<timestamp>.sidx` (sparse checkpoints)

If you pass custom `--dict` and `--post` paths, the final sparse file is generated from the dict name (same base name with `.sidx`).

## Testing

Run the engine package tests:

```bash
go test ./cmd/engine -v
```

Run all tests:

```bash
go test ./... -v
```

## Project Layout

- `main.go`: Entrypoint that delegates to the engine runner.
- `cmd/engine`: CLI parsing, mode handling, segment resolution, AND query processing.
- `internal/ingest`: Document ingestion use case and source adapters.
- `internal/index`: Searcher, analyzers, memtable and posting logic.
- `internal/store`: SSTable and sparse index serialization.
- `sample_inputs`: Example datasets for local runs.
