package kv

import (
	"github.com/stretchr/testify/require"
	"math/rand"
	"path/filepath"
	"strconv"
	"testing"
)

func BenchmarkSequentialPut(b *testing.B) {
	dir := b.TempDir()
	db, err := NewKVDB(filepath.Join(dir, "bolt.db"))

	require.NoError(b, err)
	defer db.Close()
	data := []byte("data")
	for i := 0; i < b.N; i++ {
		db.Put("test", strconv.Itoa(i), data)
	}
}

func BenchmarkRandomPut(b *testing.B) {
	dir := b.TempDir()
	db, err := NewKVDB(filepath.Join(dir, "bolt.db"))

	require.NoError(b, err)
	defer db.Close()
	data := []byte("data")
	for i := 0; i < b.N; i++ {
		val := rand.Int()
		db.Put("test", strconv.Itoa(val), data)
	}
}

func BenchmarkSequentialPutLargePayload(b *testing.B) {
	dir := b.TempDir()
	db, err := NewKVDB(filepath.Join(dir, "bolt.db"))

	require.NoError(b, err)
	defer db.Close()
	data := make([]byte, 1024*1024)
	for i := 0; i < b.N; i++ {
		db.Put("test", strconv.Itoa(i), data)
	}
}

func BenchmarkRandomPutLargePayload(b *testing.B) {
	dir := b.TempDir()
	db, err := NewKVDB(filepath.Join(dir, "bolt.db"))

	require.NoError(b, err)
	defer db.Close()
	data := make([]byte, 1024*1024)
	for i := 0; i < b.N; i++ {
		val := rand.Int()
		db.Put("test", strconv.Itoa(val), data)
	}
}

func BenchmarkBatchPut(b *testing.B) {
	dir := b.TempDir()
	db, err := NewKVDB(filepath.Join(dir, "bolt.db"))

	require.NoError(b, err)
	defer db.Close()
	data := []byte("data")
	kvs := make(map[string][]byte)
	for i := 0; i < 10_000; i++ {
		kvs[strconv.Itoa(i)] = data
	}
	for i := 0; i < b.N; i++ {
		db.PutRange("test", kvs)
	}
}

func BenchmarkGet(b *testing.B) {
	dir := b.TempDir()
	db, err := NewKVDB(filepath.Join(dir, "bolt.db"))

	require.NoError(b, err)
	defer db.Close()
	data := []byte("data")
	kvs := make(map[string][]byte)
	for i := 0; i < 10_000; i++ {
		kvs[strconv.Itoa(i)] = data
	}
	db.PutRange("test", kvs)
	for i := 0; i < b.N; i++ {
		r := rand.Int31n(10_000)
		val, found, _ := db.Get("test", strconv.Itoa(int(r)))
		require.True(b, found)
		require.True(b, string(val) == "data")
	}
}
