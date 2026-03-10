# Logging, Metrics & Distributed Tracing: The Three Pillars of Observability

## 1. Concept Overview

**Observability** is the ability to infer the internal state of a system from its external outputs. Unlike traditional monitoring (which alerts when known metrics cross thresholds), observability enables you to ask arbitrary questions about system behavior without pre-instrumenting for those specific questions.

The **Three Pillars of Observability** are:

1. **Logging** — Discrete events with contextual data (what happened, when, where)
2. **Metrics** — Numeric measurements over time (how much, how fast, how many)
3. **Tracing** — Request flow across service boundaries (how did this request propagate?)

Together, they provide complementary views: logs answer "what happened?", metrics answer "how bad is it?", and traces answer "where did it break?". Modern systems like **OpenTelemetry** unify these pillars under a single instrumentation framework.

---

## 2. Real-World Motivation

### Why Observability Matters at Scale

- **Netflix**: Processes 500B+ events/day. Without distributed tracing, debugging a failed playback across 100+ microservices would be impossible.
- **Uber**: Jaeger (open-sourced by Uber) traces millions of requests across 2000+ microservices. A single ride involves 50+ service calls.
- **Google Dapper**: The seminal 2010 paper described Google's distributed tracing system. It sampled 1 in 1024 requests yet provided full visibility—proving that sampling is viable at massive scale.
- **Amazon**: Every millisecond of latency costs revenue. Metrics (RED method) and SLOs drive continuous optimization.

### The Debugging Journey Without Observability

1. User reports: "Checkout failed"
2. Without logs: No record of the failure
3. Without metrics: No idea if it's 1 user or 10,000
4. Without traces: Must manually correlate across 20 services

With full observability: One trace ID → full request path → root cause in minutes.

---

## 3. Architecture Diagrams

### 3.1 Centralized Logging Architecture (ELK Stack)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         APPLICATION LAYER                                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │ Service A│  │ Service B│  │ Service C│  │ Service D│  │ Service E│           │
│  │ (stdout) │  │ (stdout) │  │ (stdout) │  │ (stdout) │  │ (stdout) │           │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘           │
└───────┼─────────────┼─────────────┼─────────────┼─────────────┼──────────────────┘
        │             │             │             │             │
        ▼             ▼             ▼             ▼             ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    LOG SHIPPING LAYER (per node/container)                        │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  Filebeat / Fluentd / Logstash Beats Input                                    │ │
│  │  - Tail log files or capture stdout                                           │ │
│  │  - Add metadata (hostname, pod_id, container_id)                               │ │
│  │  - Inject correlation_id from request context                                   │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
        │             │             │             │             │
        └─────────────┴─────────────┴─────────────┴─────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         LOGSTASH (optional processing)                            │
│  - Parse, transform, enrich                                                       │
│  - Route to different indices                                                     │
│  - Filter sensitive data                                                          │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         ELASTICSEARCH (storage + search)                           │
│  - Indexed by timestamp, service, level, correlation_id                            │
│  - Full-text search, aggregations                                                 │
│  - Retention policies (hot/warm/cold tiers)                                       │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         KIBANA (visualization + query)                             │
│  - Log exploration, dashboards, saved searches                                     │
│  - Correlation ID lookup → all logs for one request                               │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 Prometheus Metrics Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    INSTRUMENTED APPLICATIONS                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ /metrics     │  │ /metrics     │  │ /metrics     │  │ /metrics     │         │
│  │ :9090        │  │ :9090        │  │ :9090        │  │ :9090        │         │
│  │ Service A    │  │ Service B    │  │ Service C    │  │ Service D    │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
└─────────┼─────────────────┼─────────────────┼─────────────────┼──────────────────┘
          │                 │                 │                 │
          │   PULL (scrape)  │                 │                 │
          └─────────────────┼─────────────────┼─────────────────┘
                            │                 │
                            ▼                 ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         PROMETHEUS SERVER                                         │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  Scrape Manager (discovers targets via static config / Consul / K8s SD)     │ │
│  │  - Scrapes /metrics every 15s (configurable)                                 │ │
│  │  - Stores in local TSDB (Prometheus TSDB)                                     │ │
│  │  - PromQL query engine                                                        │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
          │                 │                 │
          ▼                 ▼                 ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────────────────────────────────┐
│   Grafana    │  │ Alertmanager │  │  Long-term storage (Thanos, Cortex, M3DB)  │
│  Dashboards  │  │  Alerting    │  │  (optional, for multi-tenancy / retention) │
└──────────────┘  └──────────────┘  └──────────────────────────────────────────┘
```

### 3.3 Distributed Tracing Architecture (OpenTelemetry + Jaeger)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         REQUEST FLOW (User → Service A → B → C)                   │
│                                                                                   │
│  User ──► [Service A] ──► [Service B] ──► [Service C] ──► DB                      │
│              │                  │                  │                              │
│              │ trace_id: abc123  │ trace_id: abc123 │ trace_id: abc123             │
│              │ span_id: s1       │ span_id: s2      │ span_id: s3                  │
│              │ parent: -         │ parent: s1       │ parent: s2                    │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                    OPENTELEMETRY SDK (in each service)                            │
│  - Creates spans for each operation                                               │
│  - Propagates trace context via headers (W3C TraceContext, B3)                     │
│  - Exports spans to OTLP collector                                                │
└─────────────────────────────────────────────────────────────────────────────────┘
          │                 │                 │
          └─────────────────┼─────────────────┘
                            │ OTLP (gRPC/HTTP)
                            ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    OPENTELEMETRY COLLECTOR                                        │
│  - Receives traces, metrics, logs                                                 │
│  - Batch processing, sampling (head-based)                                        │
│  - Exports to Jaeger, Zipkin, etc.                                                │
└─────────────────────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         JAEGER / ZIPKIN                                            │
│  - Trace storage (Cassandra, Elasticsearch, or in-memory)                         │
│  - Trace ID lookup → waterfall view                                               │
│  - Service dependency graph                                                       │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### 4.1 Logging

#### Structured Logging (JSON)

```json
{
  "timestamp": "2024-03-10T14:32:01.123Z",
  "level": "ERROR",
  "service": "checkout-service",
  "trace_id": "abc123def456",
  "span_id": "s2",
  "message": "Payment failed",
  "user_id": "u_789",
  "order_id": "ord_456",
  "error": "card_declined",
  "latency_ms": 234
}
```

**Why JSON?** Machine-parseable, easy to index, supports nested fields. Enables log aggregation and correlation.

#### Log Levels

| Level | Use Case | Production Volume |
|-------|----------|-------------------|
| DEBUG | Development, verbose flow | Rarely in prod |
| INFO | Normal operations, key events | Moderate |
| WARN | Recoverable issues, retries | Low |
| ERROR | Failures requiring attention | Low |

**Best practice**: In production, default to INFO. Use DEBUG only for specific debug sessions (dynamic log level).

#### Correlation IDs

A **correlation ID** (or **request ID**) is a unique identifier propagated through all services for a single user request. Enables:

- "Show me all logs for request X" → full request path
- Linking logs to traces (same ID in both)

Propagation: HTTP header `X-Request-ID` or `traceparent` (W3C).

### 4.2 Metrics

#### Metric Types (Prometheus)

| Type | Description | Example |
|------|-------------|---------|
| **Counter** | Monotonically increasing | `http_requests_total`, `errors_total` |
| **Gauge** | Value that can go up or down | `memory_usage_bytes`, `active_connections` |
| **Histogram** | Distribution of values + count | `http_request_duration_seconds` |
| **Summary** | Similar to histogram, client-side quantiles | `rpc_duration_seconds` |

#### RED Method (for Request-Driven Services)

- **Rate**: Requests per second
- **Errors**: Error rate (4xx, 5xx)
- **Duration**: Latency (p50, p95, p99)

#### USE Method (for Resources)

- **Utilization**: % of resource busy
- **Saturation**: Degree of queueing
- **Errors**: Error count

#### SLIs, SLOs, SLAs

| Term | Definition | Example |
|------|------------|---------|
| **SLI** | Service Level Indicator — measurable metric | "99.9% of requests < 200ms" |
| **SLO** | Service Level Objective — target for SLI | "Availability SLO: 99.9%" |
| **SLA** | Service Level Agreement — contract with consequences | "If < 99.9%, customer gets credit" |

**Error Budget**: If SLO is 99.9%, error budget = 0.1% = 43.2 minutes downtime/month. Used to gate releases.

### 4.3 Distributed Tracing

#### Concepts

- **Trace**: Full path of a request across services. Contains one or more spans.
- **Span**: Single unit of work (e.g., HTTP call, DB query). Has: trace_id, span_id, parent_span_id, name, start_time, duration, attributes.
- **Context Propagation**: trace_id + span_id passed via HTTP headers so downstream services can create child spans.

#### Sampling Strategies

| Strategy | When Decision Made | Use Case |
|----------|-------------------|----------|
| **Head-based** | At trace start | Simple, low overhead. May miss rare errors. |
| **Tail-based** | After trace completes | Sample errors, slow traces. Requires buffering. |

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| S3 durability | 11 nines (99.999999999%) |
| Elasticsearch typical retention | 7–30 days hot, 90+ days cold |
| Prometheus default scrape interval | 15 seconds |
| Prometheus TSDB retention | 15 days default |
| Jaeger sampling (Uber) | 0.1% normal, 100% errors |
| Dapper sampling (Google) | 1/1024 (0.1%) |
| Log volume (Netflix) | 500B+ events/day |
| Typical p99 latency SLO | 200–500ms for APIs |
| Error budget (99.9% SLO) | 43.2 min/month |

---

## 6. Tradeoffs (Comparison Tables)

### ELK vs Loki

| Aspect | ELK (Elasticsearch) | Loki |
|--------|---------------------|------|
| **Indexing** | Full-text index on all fields | Index only labels (like Prometheus) |
| **Log storage** | Full content indexed | Log content stored, not indexed |
| **Query** | Rich full-text, aggregations | LogQL, label-based + regex |
| **Cost** | Higher (index everything) | Lower (index less) |
| **Best for** | Complex search, compliance | Kubernetes, high volume, cost-sensitive |

### Prometheus vs InfluxDB

| Aspect | Prometheus | InfluxDB |
|--------|------------|----------|
| **Model** | Pull-based, scrape | Push-based |
| **Query** | PromQL | InfluxQL, Flux |
| **High cardinality** | Weak (labels explode memory) | Better |
| **Long-term** | Needs Thanos/Cortex | Native |
| **Use case** | Metrics, alerting | IoT, events, custom metrics |

### Jaeger vs Zipkin

| Aspect | Jaeger | Zipkin |
|--------|--------|--------|
| **Origin** | Uber (2015) | Twitter (2012) |
| **Storage** | Cassandra, ES, Kafka, memory | ES, Cassandra, MySQL |
| **UI** | Rich, dependency graph | Simpler |
| **Sampling** | Adaptive, remote config | Client-side |
| **OpenTelemetry** | Full support | Full support |

---

## 7. Variants/Implementations

### Logging

- **ELK**: Elasticsearch + Logstash + Kibana (or Beats instead of Logstash)
- **Loki + Grafana**: Log aggregation with Prometheus-style labels
- **Fluentd / Fluent Bit**: Log shippers (CNCF projects)
- **CloudWatch Logs, Stackdriver**: Cloud-native

### Metrics

- **Prometheus**: De facto standard, pull-based
- **StatsD**: Push-based, UDP, simple (counters, gauges, timers)
- **InfluxDB**: Time-series DB, push
- **Datadog, New Relic**: SaaS, full observability

### Tracing

- **Jaeger**: CNCF, production-ready
- **Zipkin**: Simpler, older
- **OpenTelemetry**: Vendor-neutral SDK + collector (traces, metrics, logs)
- **AWS X-Ray, Google Cloud Trace**: Cloud-native

---

## 8. Scaling Strategies

### Logging at Scale

1. **Sampling**: Log 1% of requests, 100% of errors
2. **Tiered retention**: Hot (7d) → warm (30d) → cold (90d) → archive
3. **Log routing**: Send only ERROR to expensive storage; INFO to cheap
4. **Async shipping**: Buffer in memory/disk, batch send

### Metrics at Scale

1. **Cardinality limits**: Avoid high-cardinality labels (user_id → bad)
2. **Recording rules**: Pre-aggregate in Prometheus
3. **Federation**: Hierarchy of Prometheus servers
4. **Thanos/Cortex**: Long-term storage, multi-tenancy

### Tracing at Scale

1. **Sampling**: 0.1%–1% of traces; 100% of errors
2. **Tail-based sampling**: Sample slow/error traces (OpenTelemetry)
3. **Trace storage**: Cassandra, Elasticsearch (sharded by trace_id)

---

## 9. Failure Scenarios

| Failure | Impact | Mitigation |
|---------|--------|------------|
| Elasticsearch down | No log search | Replication, hot standby |
| Prometheus down | No metrics, no alerts | HA pair, remote write to backup |
| Log shipper crash | Log loss on that node | Persistent queue (disk buffer) |
| High cardinality metrics | OOM in Prometheus | Limit labels, use recording rules |
| Trace storage full | Trace loss | Sampling, retention, archival |
| Correlation ID missing | Can't correlate logs | Middleware to inject in all services |

---

## 10. Performance Considerations

- **Logging**: Async I/O, avoid blocking. Structured logging has ~10–50μs overhead.
- **Metrics**: Counter increment ~100ns. Histogram observe ~1μs. Scrape every 15s keeps overhead low.
- **Tracing**: Each span ~1–5μs. Sampling reduces impact. B3 propagation adds ~100 bytes/request.

---

## 11. Use Cases

| Pillar | Use Case |
|--------|----------|
| **Logs** | Debug "why did this user's order fail?", audit trail |
| **Metrics** | "Is p99 latency increasing?", "What's our error rate?", SLO dashboards |
| **Traces** | "Which service is slow in this 10-service call?", dependency analysis |

---

## 12. Comparison Tables

### When to Use Each Pillar

| Question | Best Pillar |
|----------|-------------|
| What happened for request X? | Logs (correlation ID) |
| How fast are we? | Metrics (RED) |
| Where is the bottleneck? | Traces |
| Is the system healthy? | Metrics (SLOs) |
| Why did this user see an error? | Logs + Traces |

---

## 13. Code/Pseudocode

### Structured Logging (Python)

```python
import json
import logging

class JsonFormatter(logging.Formatter):
    def format(self, record):
        log_obj = {
            "timestamp": self.formatTime(record),
            "level": record.levelname,
            "message": record.getMessage(),
            "service": "checkout-service",
            "trace_id": getattr(record, 'trace_id', None),
        }
        if record.exc_info:
            log_obj["exception"] = self.formatException(record.exc_info)
        return json.dumps(log_obj)

# Usage
logger = logging.getLogger()
handler = logging.StreamHandler()
handler.setFormatter(JsonFormatter())
logger.addHandler(handler)
logger.info("Order created", extra={"order_id": "ord_123", "trace_id": "abc"})
```

### Prometheus Metrics (Go)

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests"},
        []string{"method", "path", "status"},
    )
    httpDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{Name: "http_request_duration_seconds", Buckets: prometheus.DefBuckets},
        []string{"method", "path"},
    )
)

func handler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    // ... handle request ...
    status := 200
    httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(status)).Inc()
    httpDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
}
```

### Distributed Tracing (OpenTelemetry + Python)

```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.jaeger.thrift import JaegerExporter

trace.set_tracer_provider(TracerProvider())
jaeger_exporter = JaegerExporter(agent_host_name="jaeger", agent_port=6831)
trace.get_tracer_provider().add_span_processor(BatchSpanProcessor(jaeger_exporter))

tracer = trace.get_tracer(__name__, "1.0.0")

def handle_request():
    with tracer.start_as_current_span("handle_request") as span:
        span.set_attribute("user.id", "u_789")
        # Call service B
        response = call_service_b()
        span.set_attribute("response.status", response.status_code)
```

### Correlation ID Middleware

```python
import uuid
from flask import g, request

def inject_correlation_id():
    g.correlation_id = request.headers.get('X-Request-ID') or str(uuid.uuid4())

def add_correlation_id_to_response(response):
    response.headers['X-Request-ID'] = g.correlation_id
    return response

# In app: before_request(inject_correlation_id), after_request(add_correlation_id)
# When calling downstream: requests.get(url, headers={'X-Request-ID': g.correlation_id})
```

---

## 14. Interview Discussion

### Key Points to Articulate

1. **Three pillars are complementary**: Logs for events, metrics for trends, traces for flow. You need all three.
2. **Correlation is critical**: Correlation ID links logs across services; trace_id links spans. Same concept.
3. **Sampling is mandatory at scale**: You cannot trace/log everything. Sample 0.1%–1%, always capture errors.
4. **SLOs drive reliability**: Error budgets prevent "perfect" from being the enemy of "shipped."
5. **OpenTelemetry is the future**: Vendor-neutral, unified instrumentation. Invest in it.

### Common Interview Questions

**Q: How would you debug a slow API that calls 5 microservices?**  
A: Start with traces—find which span is slow. Then use logs (correlation ID) for that service to see what it was doing. Metrics show if it's a trend or one-off.

**Q: Why is Prometheus pull-based?**  
A: Service discovery (Prometheus knows all targets), no firewall issues (Prometheus reaches in), simpler clients, no push storms when many instances restart.

**Q: When would you use Loki over ELK?**  
A: When log volume is huge and you mainly need to filter by labels (pod, namespace, level) rather than full-text search. Cost and operational simplicity.

**Q: How do you implement error budgets?**  
A: Define SLO (e.g., 99.9% availability). Error budget = 1 - SLO. Track actual availability. If budget exhausted, freeze releases until recovered. Google's approach.

---

## Real-World Examples (Deep Dive)

### Google Dapper (2010 Paper)

The seminal distributed tracing system. Key contributions:
- **Sampling**: 1/1024 requests traced. At Google scale, still millions of traces. Proved sampling is viable.
- **Minimal overhead**: ~1μs per span. Async collection. No impact on latency.
- **Annotation**: Application adds key-value pairs to spans (e.g., "cache hit").
- **Trace tree**: Single trace = tree of spans. Parent-child via context propagation.

### Uber Jaeger

Open-sourced 2015. Handles Uber's scale:
- **Architecture**: Agent (per host) → Collector → Storage (Cassandra/ES)
- **Sampling**: Adaptive (sample more when traffic low); remote config
- **Dependency graph**: Service A calls B, B calls C → visualize critical path
- **Integration**: OpenTracing (now OpenTelemetry)

### Netflix Atlas

Netflix's metrics system (not tracing):
- **In-memory time-series**: Sub-second resolution
- **Dimensional metrics**: Tags for filtering (region, instance, status)
- **Alerting**: Thresholds, anomaly detection
- **Scale**: Billions of time series
