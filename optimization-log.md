# Optimization Log

## 2026-06-14

Process notes:
- Stashed pre-existing dirty state, including untracked files: `stash@{0}: pre-optimization-dirty-state`.
- Benchmarks use short runs after the initial baseline: `-benchtime=200ms -count=3`.
- CPU: AMD Ryzen AI 9 365 w/ Radeon 880M, linux/amd64.
- Decision rule: keep changes that reduce allocations even if CPU time is neutral/noisy.
- Executive HTML report: `optimization-report.html`.

Clean baseline, public AVL API, 100k elements:

| Benchmark | Median ns/op | B/op | allocs/op | Notes |
| --- | ---: | ---: | ---: | --- |
| `BenchmarkTree/100000/InsertRemove` | 2050 | 1680 | 35 | Insert max key, remove it. |
| `BenchmarkTree/100000/Entries` | 2043946 | 1605635 | 1 | Allocates result slice. |
| `BenchmarkTree/100000/All` | 913502 | 0 | 0 | Full in-order traversal. |
| `BenchmarkTree/100000/All5` | 819728 | 0 | 0 | Existing benchmark used `continue`, so it was not early termination. |
| `BenchmarkTree/100000/Min` | 3.724 | 0 | 0 | Likely affected by dead-code elimination. |
| `BenchmarkTree/100000/Get` | 23.75 | 0 | 0 | Fixed key lookup. |

Clean baseline, experimental AVL-vs-B-tree comparison, 100k elements:

| Benchmark | Median ns/op | B/op | allocs/op | Notes |
| --- | ---: | ---: | ---: | --- |
| `BenchmarkComparison/100000/avl/Get` | 23.43 | 0 | 0 | Public AVL wrapper. |
| `BenchmarkComparison/100000/avl/Insert` | 883.0 | 912 | 19 | Insert max key only. |
| `BenchmarkComparison/100000/avl/Remove` | 827.2 | 753 | 15 | Remove key from original tree. |
| `BenchmarkComparison/100000/btree/Get` | 26.09 | 8 | 1 | Benchmark allocates key each op. |
| `BenchmarkComparison/100000/btree/Insert` | 852.9 | 1448 | 11 | Benchmark allocates key each op. |
| `BenchmarkComparison/100000/btree/Remove` | 943.7 | 1448 | 11 | Benchmark allocates key each op. |

### Experiment 1: benchmark harness fixes

Change:
- Added benchmark sinks so `Entries`, `All`, `Min`, `Max`, and `Get` results are consumed.
- Fixed `All5` to `break` after five elements.
- Added `Max`, `GetRandom`, and `FromMiddle5` coverage for the public AVL benchmark.

Result:
- Tests: `go test ./...` passed.
- Public AVL 100k benchmark after harness fix:

| Benchmark | Median ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkTree/100000/InsertRemove` | 2492 | 1680 | 35 |
| `BenchmarkTree/100000/Entries` | 1355434 | 1605636 | 1 |
| `BenchmarkTree/100000/All` | 737883 | 0 | 0 |
| `BenchmarkTree/100000/All5` | 43.55 | 0 | 0 |
| `BenchmarkTree/100000/FromMiddle5` | 50.27 | 0 | 0 |
| `BenchmarkTree/100000/Min` | 4.441 | 0 | 0 |
| `BenchmarkTree/100000/Max` | 3.056 | 0 | 0 |
| `BenchmarkTree/100000/Get` | 21.77 | 0 | 0 |
| `BenchmarkTree/100000/GetRandom` | 40.06 | 0 | 0 |

### Experiment 2: direct `Min`/`Max` loops

Change:
- Replaced shared `extreme(dir)` helper with direct left/right loops in `Min` and `Max`.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians from 5 runs:

| Benchmark | Baseline ns/op | Experiment ns/op | Decision |
| --- | ---: | ---: | --- |
| `BenchmarkTree/100000/Min` | 4.441 | 4.728 | Rejected |
| `BenchmarkTree/100000/Max` | 3.056 | 3.204 | Rejected |

Follow-up:
- Reverted. The shared helper is at least as good after inlining/noise.

### Experiment 3: return unchanged tree on missing `Remove`

Change:
- Added `BenchmarkTree/*/RemoveMissing`.
- Added a benchmark sink for produced tree nodes so write benchmarks consume results.
- Changed `Remove` to return the original node when the recursive child pointer is unchanged.
- Added a regression test that missing remove preserves the tree pointer.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians:

| Benchmark | Baseline ns/op | Baseline B/op | Baseline allocs/op | Experiment ns/op | Experiment B/op | Experiment allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTree/100000/RemoveMissing` | 1272 | 816 | 17 | 37.21 | 0 | 0 | Keep |

Notes:
- `InsertRemove` remains allocation-equivalent at 1680 B/op and 35 allocs/op; short-benchtime ns/op was noisy, so it is not used as evidence for this experiment.

### Experiment 4: one-pass predecessor removal

Change:
- Added `RemoveExistingMiddle` and `RemoveExistingRandom` benchmarks.
- Tried replacing `left.Max()` plus `left.Remove(max.K)` with a single recursive `removeMax` helper.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=5`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTree/100000/RemoveExistingMiddle` | 1003 | 986.1 | 768 | 16 | Rejected |
| `BenchmarkTree/100000/RemoveExistingRandom` | 1141 | 1131 | 753 | 15 | Rejected |

Follow-up:
- Reverted. The median movement is too small relative to run noise and does not improve allocations.

### Experiment 5: explicit `left`/`right` node fields

Change:
- Replaced internal `children [2]*Node` storage with explicit `left` and `right` pointers.

Result:
- Tests: `go test ./...` passed.
- Public 100k benchmark medians with `-benchtime=200ms -count=5` after the change:

| Benchmark | Baseline ns/op | Experiment ns/op | Decision |
| --- | ---: | ---: | --- |
| `BenchmarkTree/100000/Entries` | 1355434 | 1800788 | Rejected |
| `BenchmarkTree/100000/All` | 737883 | 818291 | Rejected |
| `BenchmarkTree/100000/All5` | 43.55 | 42.10 | Too small |
| `BenchmarkTree/100000/GetRandom` | 40.06 | 39.01 | Too small |

Follow-up:
- Reverted. Minor lookup/early-iteration movement does not justify slower full iteration and entries.

### Experiment 6: recursive pre-sized `Entries`

Change:
- Replaced iterative stack plus append in `Entries` with a pre-sized slice and recursive in-order fill.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=5`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTree/100000/Entries` | 1285859 | 2151638 | ~1605635 | 1 | Rejected |

Follow-up:
- Reverted. Recursive fill was much slower despite avoiding append.

### Experiment 7: cache heights inside `rotate`

Change:
- Cached child heights in local variables inside `rotate` to avoid repeated `height()` helper calls.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=5`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTree/100000/InsertRemove` | 1485 | 1719 | 1680 | 35 | Rejected |
| `BenchmarkTree/100000/RemoveMissing` | 36.77 | 39.94 | 0 | 0 | Rejected |
| `BenchmarkTree/100000/RemoveExistingMiddle` | 849.0 | 774.8 | 768 | 16 | Mixed |
| `BenchmarkTree/100000/RemoveExistingRandom` | 1025 | 992.9 | 753 | 15 | Mixed |

Follow-up:
- Reverted. Existing deletes improved modestly, but insert/remove and missing-remove slowed, allocations were unchanged, and the code got more complex.

### Experiment 8: iterative `All`

Change:
- Replaced recursive `All` traversal with an explicit stack backed by a stack-allocated array.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=5`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTree/100000/All` | 897908 | 366729 | 0 | 0 | Keep |
| `BenchmarkTree/100000/All5` | 43.22 | 17.47 | 0 | 0 | Keep |

Notes:
- The iterator still snapshots by closing over the original root pointer; the traversal mutates only a local `finger`.
- Escape check: `go test -run '^$' -gcflags='-m=2' .` reports the returned `All` func literal can escape when the `iter.Seq2` value is materialized, but `append(stack, finger) does not escape`; the explicit traversal stack stays off heap.
- Construction benchmark: storing `tree.All()` costs 24 B/op and 1 alloc/op. Direct `for range tree.All()` remains 0 B/op and 0 allocs/op.
- `NodeBuiltin.All` direct range also remains 0 B/op and 0 allocs/op; storing the returned sequence is likewise 24 B/op and 1 alloc/op.

### Experiment 9: iterative `Backward`

Change:
- Added `Backward` and `Backward5` benchmarks.
- Replaced recursive `Backward` traversal with an explicit stack backed by a stack-allocated array.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=5`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTree/100000/Backward` | 728071 | 327899 | 0 | 0 | Keep |
| `BenchmarkTree/100000/Backward5` | 39.46 | 15.42 | 0 | 0 | Keep |

### Experiment 10: iterative `From`

Change:
- Added a full `FromMiddle` benchmark.
- Replaced recursive `From` seek/traversal with an explicit stack:
  - seek pushes the path of candidate nodes with key `>= k`;
  - traversal then continues in sorted order without recursion.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=5`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTree/100000/FromMiddle5` | 49.07 | 32.48 | 0 | 0 | Keep |
| `BenchmarkTree/100000/FromMiddle` | 367096 | 141452 | 0 | 0 | Keep |

Notes:
- One `FromMiddle5` sample was an outlier; the other four experiment samples were tightly clustered around 31-33 ns/op.

### Experiment 11: iterative `BackwardFrom`

Change:
- Added `BackwardFromMiddle5` and `BackwardFromMiddle` benchmarks.
- Replaced recursive `BackwardFrom` seek/traversal with an explicit reverse stack:
  - seek pushes the path of candidate nodes with key `<= k`;
  - traversal then continues in descending order without recursion.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=5`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTree/100000/BackwardFromMiddle5` | 52.40 | 31.19 | 0 | 0 | Keep |
| `BenchmarkTree/100000/BackwardFromMiddle` | 335569 | 125301 | 0 | 0 | Keep |

Notes:
- One `BackwardFromMiddle5` sample was an outlier; the other four experiment samples were 30-36 ns/op.

### Experiment 12: exact-length `Entries` fill

Change:
- Changed `Entries` to allocate the result slice at its final length and fill by index.
- Kept the existing explicit-stack traversal, isolating the cost of `append` bookkeeping from traversal shape.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=300ms -count=5`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTree/100000/Entries` | 2066102 | 1833808 | ~1605636 | 1 | Keep |

Notes:
- Time improved by 11.2% in the focused run.
- Allocation count is unchanged because `Entries` must materialize and return the slice.

### Experiment 13: pointer-stack `Entries` traversal

Change:
- Replaced the `Entries` traversal frame state machine with the same node-pointer stack shape used by `All`.
- Kept the exact-length result slice from experiment 12.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=7`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTree/100000/Entries` | 1679869 | 1396732 | ~1605637 | 1 | Keep |

Notes:
- Time improved by 16.9% over the exact-length frame traversal.
- Allocation count is unchanged because `Entries` must materialize and return the slice.

### Experiment 14: direct `NodeBuiltin.Entries` fill

Change:
- Added `BenchmarkTreeBuiltin/*/Entries` coverage.
- Replaced `NodeBuiltin.Entries`'s generic `n.n.Entries()` call plus conversion copy with a direct in-order traversal into the final `[]Entry[K, V]`.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=7`:

| Benchmark | Baseline ns/op | Baseline B/op | Baseline allocs/op | Experiment ns/op | Experiment B/op | Experiment allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTreeBuiltin/100000/Entries` | 2350193 | 3211271 | 2 | 1612549 | 1605636 | 1 | Keep |

Notes:
- Time improved by 31.4%.
- Allocation count drops from two result-sized slices to one, and bytes/op roughly halves.

### Experiment 15: direct `NodeBuiltin.All` traversal

Change:
- Replaced `NodeBuiltin.All`'s wrapper around `n.n.All()` with a direct pointer-stack traversal that unwraps `Builtin[K]` inline.

Result:
- Focused 100k benchmark medians with `-benchtime=500ms -count=7`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTreeBuiltin/100000/All` | 340490 | 372410 | 0 | 0 | Rejected |
| `BenchmarkTreeBuiltin/100000/AllCreate` | 22.96 | 21.31 | 24 | 1 | Too small |
| `BenchmarkTreeBuiltin/100000/All5` | 15.68 | 15.60 | 0 | 0 | Too small |

Follow-up:
- Reverted. Removing the nested iterator made full traversal slower, and the tiny construction/early-iteration movement did not reduce allocations.

### Experiment 16: direct `NodeBuiltin.Get` comparisons

Change:
- Added `BenchmarkTreeBuiltin/*/Get` and `GetRandom` coverage.
- Replaced `NodeBuiltin.Get`'s `Builtin[K]` wrapper call into the generic tree with a direct loop using `<` on the built-in key type.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=7`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTreeBuiltin/100000/Get` | 22.79 | 5.016 | 0 | 0 | Keep |
| `BenchmarkTreeBuiltin/100000/GetRandom` | 40.28 | 21.06 | 0 | 0 | Keep |

Notes:
- Fixed-key lookup improved by 78.0%.
- Pseudo-random lookup improved by 47.7%.

### Experiment 17: direct `NodeBuiltin.Insert` comparisons

Change:
- Added `BenchmarkTreeBuiltin/*/InsertMax` coverage.
- Tried a specialized `NodeBuiltin.Insert` recursion using direct `<` comparisons on the built-in key type.

Result:
- Focused 100k benchmark medians with `-benchtime=500ms -count=7`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTreeBuiltin/100000/InsertMax` | 987.4 | 1100 | 912 | 19 | Rejected |

Follow-up:
- Reverted the specialized insert path. It did not reduce allocations and moved median CPU time in the wrong direction.
- Kept the benchmark coverage for future write-path experiments.

### Experiment 18: allocation-free `NodeBuiltin.Min`/`Max`

Change:
- Added `BenchmarkTreeBuiltin/*/Min` and `Max` coverage.
- Changed `NodeBuiltin.Min` and `Max` to cast the internal `*Entry[Builtin[K], V]` to `*Entry[K, V]` using `unsafe`, relying on `Builtin[K]` being a single-field wrapper around `K`.
- Added `TestMinMaxBuiltin`.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=7`:

| Benchmark | Baseline ns/op | Baseline B/op | Baseline allocs/op | Experiment ns/op | Experiment B/op | Experiment allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTreeBuiltin/100000/Min` | 22.10 | 16 | 1 | 3.105 | 0 | 0 | Keep |
| `BenchmarkTreeBuiltin/100000/Max` | 21.89 | 16 | 1 | 3.057 | 0 | 0 | Keep |

Notes:
- Allocation count drops from one heap entry copy to zero.
- This makes `NodeBuiltin.Min`/`Max` match the generic `Node.Min`/`Max` behavior of returning a pointer to the tree's internal entry.

### Experiment 19: direct `NodeBuiltin.From` traversal

Change:
- Added `BenchmarkTreeBuiltin/*/FromMiddle5` and `FromMiddle` coverage.
- Replaced `NodeBuiltin.From`'s wrapper around `n.n.From(Builtin[K]{k})` with a direct built-in seek/traversal that unwraps keys inline.
- Added `TestFromBuiltin`.

Result:
- Tests: `go test ./...` passed.
- Focused 100k benchmark medians with `-benchtime=500ms -count=7`:

| Benchmark | Baseline ns/op | Experiment ns/op | B/op | allocs/op | Decision |
| --- | ---: | ---: | ---: | ---: | --- |
| `BenchmarkTreeBuiltin/100000/FromMiddle5` | 31.64 | 17.46 | 0 | 0 | Keep |
| `BenchmarkTreeBuiltin/100000/FromMiddle` | 146063 | 142165 | 0 | 0 | Keep |

Notes:
- Five-item forward range improved by 44.8%.
- Full upper-half scan improved modestly in the focused run and remains zero-allocation.

### Post-Entries broad benchmark snapshot

Command:
- `go test -run '^$' -bench 'BenchmarkTree/100000' -benchmem -benchtime=300ms -count=3 .`

Median results from the run after experiment 19:

| Benchmark | Median ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkTree/100000/InsertRemove` | 2131 | 1680 | 35 |
| `BenchmarkTree/100000/RemoveMissing` | 36.06 | 0 | 0 |
| `BenchmarkTree/100000/RemoveExistingMiddle` | 408.6 | 768 | 16 |
| `BenchmarkTree/100000/RemoveExistingRandom` | 513.2 | 753 | 15 |
| `BenchmarkTree/100000/Entries` | 802890 | 1605638 | 1 |
| `BenchmarkTree/100000/All` | 339188 | 0 | 0 |
| `BenchmarkTree/100000/AllCreate` | 22.83 | 24 | 1 |
| `BenchmarkTree/100000/All5` | 17.46 | 0 | 0 |
| `BenchmarkTree/100000/Backward` | 323294 | 0 | 0 |
| `BenchmarkTree/100000/Backward5` | 15.74 | 0 | 0 |
| `BenchmarkTree/100000/FromMiddle5` | 30.79 | 0 | 0 |
| `BenchmarkTree/100000/FromMiddle` | 138465 | 0 | 0 |
| `BenchmarkTree/100000/BackwardFromMiddle5` | 31.98 | 0 | 0 |
| `BenchmarkTree/100000/BackwardFromMiddle` | 143595 | 0 | 0 |
| `BenchmarkTree/100000/Get` | 23.41 | 0 | 0 |
| `BenchmarkTree/100000/GetRandom` | 38.16 | 0 | 0 |
| `BenchmarkTreeBuiltin/100000/InsertMax` | 835.2 | 912 | 19 |
| `BenchmarkTreeBuiltin/100000/Entries` | 1743473 | 1605633 | 1 |
| `BenchmarkTreeBuiltin/100000/Get` | 5.487 | 0 | 0 |
| `BenchmarkTreeBuiltin/100000/GetRandom` | 21.02 | 0 | 0 |
| `BenchmarkTreeBuiltin/100000/Min` | 3.119 | 0 | 0 |
| `BenchmarkTreeBuiltin/100000/Max` | 3.019 | 0 | 0 |
| `BenchmarkTreeBuiltin/100000/All` | 456944 | 0 | 0 |
| `BenchmarkTreeBuiltin/100000/AllCreate` | 15.92 | 24 | 1 |
| `BenchmarkTreeBuiltin/100000/All5` | 15.82 | 0 | 0 |
| `BenchmarkTreeBuiltin/100000/FromMiddle5` | 16.54 | 0 | 0 |
| `BenchmarkTreeBuiltin/100000/FromMiddle` | 165759 | 0 | 0 |
