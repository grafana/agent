package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/pkg/timestamp"
	"github.com/prometheus/prometheus/tsdb/record"
	"github.com/prometheus/prometheus/tsdb/wal"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s [wal dir to read]\n", os.Args[0])
		os.Exit(1)
	}

	walDir := os.Args[1]
	if _, err := os.Stat(walDir); os.IsNotExist(err) {
		fmt.Printf("%s does not exist\n", walDir)
		os.Exit(1)
	} else if err != nil {
		fmt.Printf("error getting wal: %v\n", err)
		os.Exit(1)
	}

	// Check if /wal is a subdirectory, use that instead
	if _, err := os.Stat(filepath.Join(walDir, "wal")); err == nil {
		walDir = filepath.Join(walDir, "wal")
	}

	w, err := wal.Open(nil, prometheus.DefaultRegisterer, walDir)
	if err != nil {
		panic(err)
	}

	dir, startFrom, err := wal.LastCheckpoint(w.Dir())
	if err != nil && err != record.ErrNotFound {
		panic(err)
	}

	lookup := map[uint64]labels.Labels{}

	if err == nil {
		sr, err := wal.NewSegmentsReader(dir)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := sr.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "error closing wal segments reader: %v\n", err)
			}
		}()

		if err := printWal(wal.NewReader(sr), lookup); err != nil {
			panic(err)
		}

		startFrom++
	}

	_, last, err := w.Segments()
	if err != nil {
		panic(err)
	}

	for i := startFrom; i <= last; i++ {
		s, err := wal.OpenReadSegment(wal.SegmentName(w.Dir(), i))
		if err != nil {
			panic(err)
		}

		sr := wal.NewSegmentBufReader(s)
		err = printWal(wal.NewReader(sr), lookup)
		if err := sr.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing the wal segments reader: %v\n", err)
		}
		if err != nil {
			panic(err)
		}
	}
}

func printWal(r *wal.Reader, lookup map[uint64]labels.Labels) error {
	var dec record.Decoder

	for r.Next() {
		rec := r.Record()
		switch dec.Type(rec) {
		case record.Series:
			series, err := dec.Series(rec, nil)
			if err != nil {
				return err
			}
			for _, s := range series {
				fmt.Printf("[SERIES: %05d] %s\n", s.Ref, s.Labels.String())
				lookup[s.Ref] = s.Labels
			}
		case record.Samples:
			samples, err := dec.Samples(rec, nil)
			if err != nil {
				return err
			}
			for _, s := range samples {
				lbls, ok := lookup[s.Ref]
				if !ok {
					fmt.Printf("=== MISSING REF %d ===\n", s.Ref)
					continue
				}

				ts := timestamp.Time(s.T)

				fmt.Printf("[REF: %05d] %s\nTime: %v\tValue: %v\n\n", s.Ref, lbls.String(), ts, s.V)
			}
		default:
			return errors.New("invalid record type")
		}
	}

	return r.Err()
}
