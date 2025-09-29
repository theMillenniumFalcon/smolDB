## smolDB enhancements: fault tolerance, scalability, homepage highlights, and benchmarks

### Current state (gap analysis)
- Fault tolerance: single-process lock, no crash recovery semantics, no fsync/durability guarantees, no WAL/snapshotting, no replication.
- Scalability: single-node, directory-based index, no sharding/partitioning, no clustering or coordination, no metrics.
- Benchmarks: none. No load-gen scripts, no baseline RPS/latency, no resource profiling.
- Homepage highlights: basic features listed; no performance claims, no reliability/scalability statements, no badges/links to dashboards or benchmark reports.

---

### Goals and principles
- Keep simplicity as a core value; incrementally add reliability and scale features behind flags, without breaking current UX.
- Prefer append-only + compaction designs and filesystem primitives for durability before introducing distributed systems.
- Provide transparent, reproducible benchmarks and avoid unverifiable claims.

---

## Phase 1: Reliability and operability (single node)

1) Durability guarantees
- Add write-ahead log (WAL) for `PUT`/`PATCH`/`DELETE` operations.
  - Append JSON entries `{ts, op, key, payload}` to `wal.log` before file mutation.
  - `fsync` on WAL append; mutate JSON file; record commit entry; optional group commit knob.
  - On startup, replay WAL and reconcile with index.
- Add `fsync`/`O_DSYNC` controls and configurable durability levels: `none`, `commit`, `grouped(ms)`.

2) Crash recovery
- On boot: verify lock status, rebuild index, replay WAL, truncate at last good entry, then checkpoint.
- Introduce periodic snapshots/checkpoints: `checkpoint/000NNN.snap` with compacted state and truncated WAL.

3) Data integrity
- Add per-file checksum (e.g., xxhash) stored alongside `<key>.json.meta` or embedded header.
- Validate checksum on read; expose `GET /integrity/:key` and `POST /integrity/repair`.

4) Observability
- Structured logs with request IDs; add Prometheus metrics: qps, latency histograms, errors, in-flight, WAL queue depth, fsync time.
- Add `/metrics` endpoint.

5) Operational tooling
- Offline compaction command: `smoldb admin compact` to rewrite JSON and trim WAL.
- Integrity scanner: `smoldb admin verify` to scan all keys, report or repair.

Acceptance criteria
- No data loss on power-cut tests under `commit` durability.
- Clean restart with WAL replay and consistent index.
- Metrics visible in Prometheus + basic Grafana dashboard JSON committed.

---

## Phase 2: Horizontal scalability options

A) Read scaling via replicas (eventual consistency)
- Add asynchronous follower replicas with pull-based log shipping.
- New components:
  - Leader exposes `/replication/stream` (cursor-based) delivering WAL segments.
  - Follower `smoldb replicate --from <leader>` applies WAL, maintains lag metrics.
- Client routing: document in README to place a TCP/HTTP load balancer for GETs; writes go to leader.

B) Partitioning (sharding) alpha
- Hash keys to N shards; each shard is a `FileIndex` under subdirectories.
- Router inside server: `shard = hash(key) % N`.
- Start with single-process multi-shard (N>=1). Future: distribute shards across nodes via static map.

C) Consistency model
- Document: reads are eventually consistent on replicas, strongly consistent on leader; writes are linearized per key.
- Provide `consistency=leader|replica_ok` query param for GETs when behind a proxy that supports node selection.

Acceptance criteria
- Replica can trail and catch up; consistent replay on restarts; observability on lag.
- Shard router correctness; no cross-shard deadlocks; performance improves for mixed keys.

---

## Phase 3: Homepage highlights and documentation

- Features section
  - Key-value JSON store with O(1) index lookup
  - Reference resolution with configurable depth
  - Durability levels and WAL-based crash recovery
  - Read replicas (beta) and shardable storage (alpha)
  - Observability: Prometheus metrics, Grafana dashboard
  - Simple CLI+Docker deploy

- Reliability/scale claims
  - Clearly state tested durability level and environment.
  - State replica consistency model and shard router behavior.

- Benchmarks section
  - Link to reproducible bench scripts and raw results.
  - Present latency P50/P95/P99 and throughput; CPU/memory IO stats.

- Badges and links
  - GitHub Actions test badge, Go version badge, Docker pulls, Benchmark report link, Grafana dashboard JSON link.

---

## Phase 4: Benchmarking plan (reproducible)

Workloads
- GET key hit, PUT overwrite, PATCH field, DELETE, 80/20 read/write mix, and REF resolution depths 0/1/3.

Environments
- Local: `docker compose` single-node and leader+replica.
- Cloud: c6i.large (2 vCPU, 4 GiB), gp3 SSD 2000 IOPS baseline.

Tools
- `hey` for HTTP load, `bombardier` alternative, `k6` for scripted scenarios.
- Collect: Prometheus metrics + `pidstat`/`iostat`/`perf stat`.

Methodology
- Warmup 60s, measure 120s, cool down 30s; three runs per scenario.
- Control for GC (set GOGC), file system cache (drop caches where possible), durability level.

Reporting
- Store raw JSON/CSV under `bench/results/<date>`; generate markdown report in `bench/reports` with plots.
- Include environment details, versions, exact commands.

Acceptance criteria
- Scripts are runnable via `make bench`.
- CI runs a smoke benchmark on PR (low load).

---

## Phase 5: Milestones, tasks, and sequencing

M1: Durability + recovery (2-3 weeks)
- Implement WAL append + replay; fsync knobs; checkpoints; tests for crash-recovery; docs.

M2: Observability + ops (1 week)
- Prometheus metrics; Grafana dashboard JSON; admin commands; docs.

M3: Read replicas (2 weeks)
- WAL streaming API; follower process; lag metrics; fail/restart tests; docs.

M4: Sharding alpha (1-2 weeks)
- In-process shard router + multiple `FileIndex` instances; migration tool; docs.

M5: Benchmarks + homepage refresh (1 week)
- Scripts, reports, updated README/web highlights and badges.

---

## Technical implementation notes

WAL format (line-delimited JSON)
```json
{"v":1,"ts":1710000000,"op":"PUT","key":"k","body":"...raw..."}
{"v":1,"ts":1710000001,"op":"PATCH","key":"k","field":"f","body":"..."}
{"v":1,"ts":1710000002,"op":"DELETE","key":"k"}
```
- Append atomically; fsync after N ms (configurable); use file rotate by size/time.
- On replay, validate sequence and checksums; stop at first malformed line.

Prometheus metrics
- `smoldb_http_requests_total{route,method,status}`
- `smoldb_http_request_duration_seconds_bucket{route}`
- `smoldb_wal_appends_total`, `smoldb_wal_fsync_seconds_bucket`
- `smoldb_index_size`, `smoldb_keys_total`
- `smoldb_replica_lag_seconds` (when enabled)

Config surface
- CLI flags/env: `--dir`, `--port`, `--durability`, `--group-commit-ms`, `--metrics`.

Testing
- Fault-injection tests: kill -9 mid-write; power-cut simulation; replay idempotency; checksum corruption.
- Load tests in CI (reduced scale) for regression detection.

Migration
- Backward compatible with current `.json` layout; WAL and meta files live in hidden subdir `.smoldb/`.

Risks
- WAL fsync can cap throughput; group commit mitigates.
- Reference resolution cost at high depth; mitigate via caching or depth limits.


