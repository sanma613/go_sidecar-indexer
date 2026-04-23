package engine

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"practice/indexing-motor/internal/core"
	"practice/indexing-motor/internal/index"
	"practice/indexing-motor/internal/ingest"
	indexeradapter "practice/indexing-motor/internal/ingest/adapters/indexer"
	sourceadapter "practice/indexing-motor/internal/ingest/adapters/source"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	format           string
	inputPath        string
	inputDir         string
	dictPath         string
	postPath         string
	segmentDir       string
	mode             string
	query            string
	checkpointStride uint64
	memThreshold     int64
	memHardLimit     int64
}

func Run(args []string) error {
	config, err := parseConfig(args)
	if err != nil {
		return err
	}

	if config.mode == "ingest" || config.mode == "ingest-search" {
		source, err := sourceadapter.NewDocumentSource(config.format, config.inputPath, config.inputDir)
		if err != nil {
			return err
		}

		indexer := indexeradapter.NewMemTableIndexerWithLimits(config.checkpointStride, config.memThreshold, config.memHardLimit)
		useCase := ingest.NewUseCase(source, indexer)

		dictPath := config.dictPath
		postPath := config.postPath
		if dictPath == "" || postPath == "" {
			timestamp := fmt.Sprintf("%016d", time.Now().UnixNano())
			dictPath = filepath.Join(config.segmentDir, "dict-"+timestamp+".sst")
			postPath = filepath.Join(config.segmentDir, "post-"+timestamp+".sst")
		}

		documentCount, err := useCase.Execute(context.Background(), dictPath, postPath)
		if err != nil {
			return err
		}

		fmt.Printf("ingested %d documents\n", documentCount)
		fmt.Printf("segment: %s | %s\n", dictPath, postPath)
	}

	if config.mode == "search" || config.mode == "ingest-search" {
		if strings.TrimSpace(config.query) == "" {
			return errors.New("query is required in search mode")
		}

		terms := parseQueryTerms(config.query)
		if len(terms) == 0 {
			return errors.New("query is required in search mode")
		}

		segments, err := resolveSearchSegments(config)
		if err != nil {
			return err
		}
		if len(segments) == 0 {
			return errors.New("no segments found to search")
		}

		fmt.Printf("results for %q:\n", config.query)
		total := 0
		for _, segment := range segments {
			searcher, openErr := index.NewSearcher(segment.dictPath, segment.postPath)
			if openErr != nil {
				continue
			}

			resultIDs, searchErr := searchByTermsAND(searcher, terms)
			if searchErr != nil {
				continue
			}

			for _, docID := range resultIDs {
				fmt.Printf("%s -> %d\n", segment.name, docID)
				total++
			}
		}

		if total == 0 {
			fmt.Println("no results")
		}
	}

	return nil
}

func parseConfig(args []string) (Config, error) {
	fs := flag.NewFlagSet("engine", flag.ContinueOnError)
	format := fs.String("format", "auto", "input format: auto|csv|json|pdf")
	inputPath := fs.String("input", "", "input file path")
	inputDir := fs.String("input-dir", "", "input directory path (parse all csv/json/pdf files)")
	dictPath := fs.String("dict", "", "dictionary sstable output path (optional, default: timestamp segment)")
	postPath := fs.String("post", "", "postings sstable output path (optional, default: timestamp segment)")
	segmentDir := fs.String("segment-dir", "data", "segment directory for timestamped sst files")
	mode := fs.String("mode", "ingest", "mode: ingest|search|ingest-search")
	query := fs.String("query", "", "query text for search mode")
	checkpointStride := fs.Uint64("checkpoint-stride", 64, "sparse index checkpoint stride")
	memThreshold := fs.Int64("mem-threshold-bytes", core.MemTableThreshold, "memtable flush threshold in bytes")
	memHardLimit := fs.Int64("mem-hard-limit-bytes", core.MemTableHardLimit, "memtable hard limit in bytes")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	config := Config{
		format:           strings.ToLower(strings.TrimSpace(*format)),
		inputPath:        strings.TrimSpace(*inputPath),
		inputDir:         strings.TrimSpace(*inputDir),
		dictPath:         strings.TrimSpace(*dictPath),
		postPath:         strings.TrimSpace(*postPath),
		segmentDir:       strings.TrimSpace(*segmentDir),
		mode:             strings.ToLower(strings.TrimSpace(*mode)),
		query:            strings.TrimSpace(*query),
		checkpointStride: *checkpointStride,
		memThreshold:     *memThreshold,
		memHardLimit:     *memHardLimit,
	}

	if config.mode != "ingest" && config.mode != "search" && config.mode != "ingest-search" {
		return Config{}, fmt.Errorf("invalid mode %q", config.mode)
	}

	if (config.mode == "ingest" || config.mode == "ingest-search") && config.inputPath == "" && config.inputDir == "" {
		return Config{}, errors.New("input or input-dir is required in ingest mode")
	}

	if (config.dictPath == "") != (config.postPath == "") {
		return Config{}, errors.New("dict and post must be provided together")
	}

	if config.segmentDir == "" {
		return Config{}, errors.New("segment-dir is required")
	}

	if config.memThreshold <= 0 {
		return Config{}, errors.New("mem-threshold-bytes must be > 0")
	}
	if config.memHardLimit <= config.memThreshold {
		return Config{}, errors.New("mem-hard-limit-bytes must be greater than mem-threshold-bytes")
	}

	return config, nil
}

type segmentPair struct {
	name     string
	dictPath string
	postPath string
	ts       int64
}

func resolveSearchSegments(config Config) ([]segmentPair, error) {
	if config.dictPath != "" && config.postPath != "" {
		return []segmentPair{{
			name:     filepath.Base(config.dictPath),
			dictPath: config.dictPath,
			postPath: config.postPath,
			ts:       0,
		}}, nil
	}

	entries, err := os.ReadDir(config.segmentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []segmentPair{}, nil
		}
		return nil, err
	}

	segments := make([]segmentPair, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "dict-") || !strings.HasSuffix(name, ".sst") {
			continue
		}

		token := strings.TrimSuffix(strings.TrimPrefix(name, "dict-"), ".sst")
		timestamp, parseErr := strconv.ParseInt(token, 10, 64)
		if parseErr != nil {
			continue
		}

		dictPath := filepath.Join(config.segmentDir, name)
		postPath := filepath.Join(config.segmentDir, "post-"+token+".sst")
		if _, statErr := os.Stat(postPath); statErr != nil {
			continue
		}

		segments = append(segments, segmentPair{
			name:     token,
			dictPath: dictPath,
			postPath: postPath,
			ts:       timestamp,
		})
	}

	sort.Slice(segments, func(i, j int) bool {
		return segments[i].ts > segments[j].ts
	})

	return segments, nil
}

func parseQueryTerms(query string) []string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return nil
	}

	rawParts := strings.Split(normalized, " and ")
	terms := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		term := strings.TrimSpace(part)
		if term == "" {
			continue
		}
		terms = append(terms, term)
	}

	return terms
}

func searchByTermsAND(searcher *index.Searcher, terms []string) ([]uint64, error) {
	if len(terms) == 0 {
		return []uint64{}, nil
	}

	termDocIDs := make([][]uint64, 0, len(terms))
	for _, term := range terms {
		token := []byte(term)
		index.LowerCaseFilter(token)
		h1, h2 := core.DoubleHash(token)

		postings, err := searcher.Search(core.NewInvertedMemTableKey(h1, h2))
		if err != nil {
			if errors.Is(err, index.ErrKeyNotFound) {
				return []uint64{}, nil
			}
			return nil, err
		}

		docIDs := make([]uint64, 0, 32)
		for postings.Next() {
			docIDs = append(docIDs, postings.Value())
		}
		if postings.Err() != nil {
			return nil, postings.Err()
		}
		termDocIDs = append(termDocIDs, docIDs)
	}

	slices.SortFunc(termDocIDs, func(a, b []uint64) int {
		switch {
		case len(a) < len(b):
			return -1
		case len(a) > len(b):
			return 1
		default:
			return 0
		}
	})

	result := append([]uint64(nil), termDocIDs[0]...)
	for i := 1; i < len(termDocIDs); i++ {
		result = intersectSortedDocIDs(result, termDocIDs[i])
		if len(result) == 0 {
			return result, nil
		}
	}

	return result, nil
}

func intersectSortedDocIDs(a, b []uint64) []uint64 {
	result := make([]uint64, 0, min(len(a), len(b)))
	i := 0
	j := 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			result = append(result, a[i])
			i++
			j++
		case a[i] < b[j]:
			i++
		default:
			j++
		}
	}

	return result
}
