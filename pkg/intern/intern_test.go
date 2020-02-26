// This file is modified from the interner found in Prometheus. The original
// license follows:
//
// Copyright 2019 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Inspired / copied / modified from https://gitlab.com/cznic/strutil/blob/master/strutil.go,
// which is MIT licensed, so:
//
// Copyright (c) 2014 The strutil Authors. All rights reserved.

package intern

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIntern(t *testing.T) {
	interner := New().(*pool)
	testString := "TestIntern"
	interner.Intern(testString)
	interned, ok := interner.pool[testString]

	require.True(t, ok)
	require.Equal(t, int64(1), interned.refs)
}

func TestIntern_MultiRef(t *testing.T) {
	interner := New().(*pool)
	testString := "TestIntern_MultiRef"

	interner.Intern(testString)
	interned, ok := interner.pool[testString]

	require.True(t, ok)
	require.Equal(t, int64(1), interned.refs)

	interner.Intern(testString)
	interned, ok = interner.pool[testString]

	require.True(t, ok)
	require.Equal(t, int64(2), interned.refs)
}

func TestIntern_DeleteRef(t *testing.T) {
	interner := New().(*pool)
	testString := "TestIntern_DeleteRef"

	interner.Intern(testString)
	interned, ok := interner.pool[testString]

	require.True(t, ok)
	require.Equal(t, int64(1), interned.refs)

	interner.Release(testString)
	_, ok = interner.pool[testString]
	require.False(t, ok)
}

func TestIntern_MultiRef_Concurrent(t *testing.T) {
	interner := New().(*pool)
	testString := "TestIntern_MultiRef_Concurrent"

	interner.Intern(testString)
	interned, ok := interner.pool[testString]
	require.True(t, ok)
	require.Equal(t, int64(1), interned.refs)

	go interner.Release(testString)

	interner.Intern(testString)

	time.Sleep(time.Millisecond)

	interner.mtx.RLock()
	interned, ok = interner.pool[testString]
	interner.mtx.RUnlock()
	require.True(t, ok)
	require.Equal(t, int64(1), atomic.LoadInt64(&interned.refs))
}
