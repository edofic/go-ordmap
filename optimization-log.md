# Optimization Log

## 2026-06-14

Process notes:
- Stashed pre-existing dirty state, including untracked files: `stash@{0}: pre-optimization-dirty-state`.
- Benchmarks use short runs after the initial baseline: `-benchtime=200ms -count=3`.
- CPU: AMD Ryzen AI 9 365 w/ Radeon 880M, linux/amd64.

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
