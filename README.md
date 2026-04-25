# Go Sidecar Indexer

![Go Version](https://img.shields.io/badge/Go-1.23-blue)
![Status](https://img.shields.io/badge/status-active-brightgreen)

Go Sidecar Indexer is a lightweight, high-performance full-text indexing engine for CSV, JSON, and PDF document collections. It’s designed for rapid ingestion, immutable on-disk segment storage (SSTables), and efficient AND-term search with minimal memory overhead.

---

## 🚀 Why Use Go Sidecar Indexer?

- **Professionally engineered** for fast, scalable document indexing and search.
- Ideal for building search layers into microservices, ETL/data pipelines, or for research/computing use cases.
- Excels with heterogeneous file formats (CSV, JSON, PDF) with zero schema config.
- Scriptable CLI—suitable for production piloting and integration into data-heavy microservices.
- **Written in Go, using efficient, modern systems patterns.**

---

## Technologies Used

- **Language:** Go 1.23
- **Key Design Patterns:** Segment-based inverted index, immutable SSTables, sparse checkpointing
- **Dependencies:** Minimal, using support utilities from `golang.org/x/exp`
- **Demo/Test Setup:** Fast local sample file ingestion and robust AND-term searches
- **Professional Application:** Under-the-hood techniques similar to LSM-trees and modern search engines

---

## Features

- Ingest documents from a single file or directory: CSV, JSON, PDF (`auto` detection)
- Builds timestamped, immutable on-disk segments:  
  - Dictionary (`dict-*.sst`)  
  - Postings (`post-*.sst`)
- Sparse index checkpoints for efficient partial loads
- Term-based AND queries (all terms must match, normalized to lowercase)
- **Smart Search Defaults:** You don't have to specify segment files—if not given, the engine auto-discovers and queries the latest available segments in the `data/` directory.
- Bundled CLI demo and realistic sample data for immediate hands-on experimentation

---

## Quick Start

### 1. Run ingest mode with a single file

```bash
go run . --mode ingest --input sample_inputs/products.json
```

### 2. Ingest all supported files in a directory

```bash
go run . --mode ingest --input-dir sample_inputs
```

### 3. Run search mode

```bash
go run . --mode search --query "machine learning and optimization"
```
> **Tip:** You do **not** need to specify `--dict` and `--post` unless you want to search a specific segment pair. By default, the engine searches all the latest segments in `data/`.

### 4. Search with explicit segment pair

```bash
go run . --mode search \
  --query "deep learning" \
  --dict data/dict-XXXXXXXXX.sst \
  --post data/post-XXXXXXXXX.sst
```

### 5. Ingest + search in one command

```bash
go run . --mode ingest-search --input-dir sample_inputs --query "AI agent"
```

### 6. Build and run a binary

```bash
go build -o sidecar-indexer .
./sidecar-indexer --mode ingest --input-dir sample_inputs
```

### 7. One-command demo

```bash
bash run_hex_demo.sh
```

---

## CLI Flags

| Flag                     | Default            | Description                                                      |
|--------------------------|--------------------|------------------------------------------------------------------|
| `--mode`                 | `ingest`           | Execution mode: `ingest`, `search`, `ingest-search`              |
| `--format`               | `auto`             | Input parser: `auto`, `csv`, `json`, `pdf`                       |
| `--input`                | `""`               | Input file path                                                  |
| `--input-dir`            | `""`               | Input directory (scan csv/json/pdf)                              |
| `--dict`                 | `""`               | Dict SST file path (optional, use with `--post`)                 |
| `--post`                 | `""`               | Postings SST file path (optional, use with `--dict`)             |
| `--segment-dir`          | `data`             | Output directory for timestamped segments                        |
| `--query`                | `""`               | Query string for search modes                                    |
| `--checkpoint-stride`    | `64`               | Sparse checkpoint interval                                       |
| `--mem-threshold-bytes`  | `54 * 1024 * 1024` | Memtable flush threshold                                         |
| `--mem-hard-limit-bytes` | `70 * 1024 * 1024` | Memtable hard limit (must exceed threshold)                      |

---

## Query Behavior

- Terms are normalized to lowercase.
- AND is supported using the string `and` (case-insensitive).
- Example: `"go AND search and engine"`.

---

## Project Structure

- `main.go`: Entrypoint for the application.
- `cmd/engine`: CLI parsing, mode handling, segment/file management, search workflow.
- `internal/ingest`: Document source adapters and ingestion logic.
- `internal/index`: Search algorithms, analyzers, posting/memtable logic.
- `internal/store`: SSTable and sparse index serialization.
- `sample_inputs/`: Realistic datasets for demonstrations and tests.

---

## Sample Data

The `sample_inputs/` directory contains realistic CSV, JSON, and PDF datasets for demonstration and integration tests with diverse formats.

---

## Testing

- Run package-specific tests:
  ```bash
  go test ./cmd/engine -v
  ```
- Run all tests:
  ```bash
  go test ./... -v
  ```

---

Maintained by Santiago Machado. Open to technical discussions, PRs, and collaborations regarding search engine optimization and AI-driven data systems.
