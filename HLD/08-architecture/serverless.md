# Serverless Architecture

## 1. Concept Overview

### Definition
Serverless architecture is a cloud-native development model where the cloud provider dynamically manages the allocation and provisioning of servers. Developers write and deploy code without managing server infrastructure. The term encompasses **FaaS (Function as a Service)**—event-driven, ephemeral function execution—and **BaaS (Backend as a Service)**—managed backend services like databases, auth, and storage.

### Purpose
- **Zero server management**: No provisioning, patching, or scaling of servers
- **Pay-per-use**: Charge only for actual execution time and resources consumed
- **Automatic scaling**: Scale to zero when idle; scale out automatically under load
- **Rapid development**: Deploy code without infrastructure setup
- **Reduced operational burden**: Focus on business logic, not infrastructure

### Problems It Solves
- **Over-provisioning**: Pay for idle capacity in traditional models
- **Under-provisioning**: Avoid capacity planning errors during traffic spikes
- **Operational overhead**: Eliminate server maintenance, OS patches, scaling config
- **Development velocity**: Faster iteration without DevOps bottleneck
- **Cost predictability**: Pay only for what you use (with caveats)

---

## 2. Real-World Motivation

### Netflix
- **Media encoding**: Lambda for video transcoding pipelines; triggers on S3 upload
- **Chaos engineering**: Lambda for automated chaos experiments (Chaos Monkey, etc.)
- **Data processing**: Event-driven ETL; Lambda processes events from Kinesis/SQS
- **Scale**: Thousands of Lambda invocations per second during encoding jobs
- **Why serverless**: Bursty, event-driven workloads; don't need always-on servers

### Coca-Cola
- **Vending machine APIs**: 16M+ vending machines; Lambda handles API requests
- **Scale**: Millions of requests; unpredictable geographic distribution
- **Benefit**: Scale to zero when machines idle; pay per interaction
- **Use case**: Machine status, inventory, payment processing

### iRobot (Roomba)
- **IoT data ingestion**: Lambda processes data from millions of connected devices
- **Event-driven**: Device events trigger processing; no constant load
- **Cost**: Significant savings vs. always-on EC2 fleet

### Slack
- **Message processing**: Lambda for real-time message indexing, search
- **Event-driven**: Message events trigger indexing pipelines
- **Burst handling**: Handles traffic spikes during peak hours

### Nordstrom
- **Image processing**: Lambda for resizing, optimization on S3 upload
- **E-commerce**: Product image pipeline; event-driven on upload

---

## 3. Architecture Diagrams

### FaaS Execution Model

```
                    ┌─────────────────────────────────────────┐
                    │           CLOUD PROVIDER                  │
                    │  ┌─────────────────────────────────────┐ │
                    │  │         REQUEST ROUTER               │ │
                    │  └─────────────────┬───────────────────┘ │
                    │                    │                      │
                    │    ┌───────────────┼───────────────┐      │
                    │    │               │               │      │
                    │    ▼               ▼               ▼      │
                    │ ┌──────┐       ┌──────┐       ┌──────┐   │
                    │ │ Fn 1 │       │ Fn 2 │       │ Fn N │   │
                    │ │(cold)│       │(warm)│       │(warm)│   │
                    │ └──────┘       └──────┘       └──────┘   │
                    │    │               │               │      │
                    │    └───────────────┼───────────────┘      │
                    │                    │                      │
                    │            Auto-scaling (0 → N)            │
                    └─────────────────────────────────────────┘
```

### Event Triggers and FaaS

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   S3       │     │   API       │     │   SQS       │
│  Upload    │     │  Gateway    │     │   Queue     │
└──────┬─────┘     └──────┬──────┘     └──────┬──────┘
       │                  │                   │
       │                  │                   │
       └──────────────────┼───────────────────┘
                          │
                          ▼
                 ┌─────────────────┐
                 │  AWS LAMBDA      │
                 │  (or equivalent) │
                 └────────┬─────────┘
                          │
       ┌──────────────────┼──────────────────┐
       ▼                  ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   DynamoDB  │    │     S3      │    │   External  │
│   (BaaS)    │    │   (BaaS)    │    │     API     │
└─────────────┘    └─────────────┘    └─────────────┘
```

### Cold Start vs Warm Start

```
COLD START (First request or after idle)
────────────────────────────────────────
Request ──▶ [Init runtime] ──▶ [Load code] ──▶ [Run handler] ──▶ Response
              ~100-500ms          ~50-200ms       ~10-100ms
              Total: 200ms - 2s+ (language/runtime dependent)

WARM START (Reused container)
────────────────────────────────────────
Request ──▶ [Run handler] ──▶ Response
              ~1-50ms
```

---

## 4. Core Mechanics

### FaaS Execution Model
1. **Event occurs**: HTTP request, S3 upload, SQS message, scheduled trigger
2. **Platform routes** to available container or spins new one
3. **Runtime initialized** (cold) or reused (warm)
4. **Handler invoked** with event payload
5. **Response returned** to caller or event source
6. **Container kept warm** for reuse (platform-dependent timeout, typically 5-15 min)

### Stateless Functions
- No in-memory state between invocations
- Persist state in external store (DynamoDB, S3, Redis)
- Each invocation may run in different container
- Design for idempotency (retries can duplicate invocations)

### Event Triggers
- **HTTP**: API Gateway, ALB
- **Storage**: S3 (object create, delete), DynamoDB Streams
- **Messaging**: SQS, SNS, Kafka (via connector)
- **Scheduled**: CloudWatch Events, cron
- **Database**: RDS Proxy, Aurora Serverless (different model)

### Scaling Model
- **Automatic**: Platform scales instances based on incoming events
- **Per-request**: Each request can get dedicated instance (concurrency = 1) or share (concurrency > 1)
- **Scale to zero**: No cost when no invocations
- **Burst limits**: AWS Lambda 1000 concurrent (default), 3000 burst

---

## 5. Numbers

| Metric | AWS Lambda | Google Cloud Functions | Azure Functions |
|--------|-------------|------------------------|-----------------|
| Max timeout | 15 min | 9 min (gen1), 60 min (gen2) | 10 min (consumption) |
| Max memory | 10 GB | 8 GB | 1.5 GB (consumption) |
| Max payload (sync) | 6 MB | 10 MB | 100 KB (HTTP) |
| Cold start (Node) | 50-200ms | 100-400ms | 200-500ms |
| Cold start (Java) | 1-5s | 2-6s | 2-5s |
| Free tier (AWS) | 1M req/mo, 400K GB-sec | 2M invocations | 1M executions |

### Cold Start by Runtime (AWS Lambda, approximate)

| Runtime | Cold Start | Warm Start |
|---------|------------|------------|
| Node.js | 50-150ms | 1-5ms |
| Python | 100-300ms | 1-5ms |
| Go | 50-200ms | 1-3ms |
| Java | 1-5s | 10-50ms |
| .NET | 500ms-2s | 5-20ms |
| Rust (custom) | 20-50ms | <1ms |

### Cost Model (AWS Lambda, us-east-1)
- **Requests**: $0.20 per 1M requests
- **Duration**: $0.0000166667 per GB-second
- **Example**: 1M requests, 256MB, 200ms each = $0.20 + $0.33 = $0.53

---

## 6. Tradeoffs

### Serverless vs Traditional (EC2/Containers)

| Aspect | Serverless | Traditional |
|--------|------------|-------------|
| Cold start | Yes (latency spike) | No |
| Max execution time | 15 min (Lambda) | Unlimited |
| State | Stateless only | Can hold state |
| Scaling | Automatic, fine-grained | Manual or auto-scaling groups |
| Cost (low traffic) | Very low (pay per use) | Pay for idle |
| Cost (high traffic) | Can exceed EC2 at scale | Often cheaper |
| Vendor lock-in | High | Lower (containers portable) |
| Debugging | Harder (distributed) | Easier (single process) |
| Local dev | Emulation needed | Run natively |

### When Serverless vs When to Avoid

| Use Serverless | Avoid Serverless |
|----------------|------------------|
| Event-driven (S3, SQS, webhooks) | Long-running processes (>15 min) |
| Sporadic traffic | Steady high traffic |
| Rapid prototyping | Complex stateful workflows |
| Glue code (ETL, orchestration) | WebSockets (connection state) |
| Scheduled jobs | Need for persistent connections |
| Variable load | Predictable, constant load |
| Cost optimization (low traffic) | Cost optimization (high traffic) |

---

## 7. Variants / Implementations

### AWS Lambda
- **Triggers**: API Gateway, S3, SQS, SNS, DynamoDB Streams, EventBridge, etc.
- **Runtimes**: Node, Python, Java, Go, .NET, Ruby, custom (container)
- **Provisioned Concurrency**: Pre-warm to eliminate cold starts (extra cost)
- **Lambda@Edge**: Run at CDN edge (CloudFront) for low latency

### Google Cloud Functions
- **Gen1**: Event-driven; Gen2: HTTP + event-driven, longer timeout
- **Runtimes**: Node, Python, Go, Java, .NET, Ruby, PHP
- **Cloud Run**: Container-based serverless; more control, no 15-min limit

### Azure Functions
- **Consumption plan**: True serverless; scale to zero
- **Premium plan**: Pre-warmed instances; VNet integration
- **Durable Functions**: Stateful workflows (orchestrator pattern)

### Cloudflare Workers
- **Edge execution**: Runs at 275+ locations globally
- **V8 isolates**: Sub-millisecond cold starts
- **Limits**: 50ms CPU (free), 30s (paid)
- **Use case**: Edge logic, A/B testing, geo-routing

### Alibaba Cloud Function Compute
- Similar to Lambda; strong in China market

---

## 8. Scaling Strategies

### Cold Start Mitigation
- **Provisioned Concurrency** (AWS): Pre-initialize N instances; eliminates cold start for those
- **Keep-warm**: Scheduled ping every 5 min to prevent idle timeout
- **Smaller runtimes**: Node, Python, Go have faster cold starts than Java
- **Custom runtime**: Minimize init (e.g., Rust, Go compiled binary)
- **Reserve concurrency**: Guarantee minimum capacity

### Performance Optimization
- **Increase memory**: More CPU allocated; can reduce duration (cost trade-off)
- **Connection pooling**: Reuse DB connections across invocations (warm container)
- **Lazy initialization**: Initialize clients outside handler (reused in warm)
- **Minimize package size**: Faster cold start; tree-shake unused deps

### Cost Optimization
- **Right-size memory**: Profile; often 512MB-1GB sufficient
- **Reduce duration**: Optimize code; avoid unnecessary work
- **Batch processing**: Process SQS messages in batches (up to 10)
- **Reserved capacity**: For predictable load; provisioned concurrency

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Cold start timeout | Client timeout | Provisioned concurrency; keep-warm |
| Throttling | 429; dropped events | Reserved concurrency; SQS for async |
| Out of memory | Function killed | Increase memory; fix memory leaks |
| Timeout | Partial execution | Idempotency; checkpoint long work |
| Concurrent limit | New invocations queued | Request increase; use SQS buffer |
| Region outage | Functions unavailable | Multi-region; failover |
| Vendor lock-in | Migration cost | Abstract with framework (Serverless Framework) |

### Retry Behavior
- **Async invocations** (SQS, S3): Automatic retries (2-6 times)
- **Sync invocations** (API Gateway): No automatic retry; client must retry
- **Dead letter queue**: Send failed events for manual handling

---

## 10. Performance Considerations

- **Cold start**: Dominates P99 latency for infrequent functions
- **Network**: Each function is network hop; minimize cross-function calls
- **Connection limits**: RDS has connection limits; use RDS Proxy or connection pooling
- **Payload size**: Large payloads add latency; use S3 for large data
- **Concurrency**: High concurrency can exhaust downstream (DB, API) connections
- **Billing**: Duration rounded up to nearest 1ms; optimize hot path

---

## 11. Use Cases

**Ideal:**
- Webhook handlers
- Image/video processing (on upload)
- Scheduled data processing
- API backends (low-medium traffic)
- Event-driven ETL
- IoT data ingestion
- Chatbots, Alexa skills

**Poor fit:**
- WebSocket servers (connection state)
- Long-running batch jobs (>15 min)
- High-throughput, low-latency APIs (cold start)
- Monolithic migration (without decomposition)
- Applications requiring persistent in-memory state

---

## 12. Comparison Tables

### FaaS Provider Comparison

| Feature | AWS Lambda | GCP Functions | Azure Functions |
|---------|------------|---------------|-----------------|
| Max timeout | 15 min | 60 min (gen2) | 10 min (consumption) |
| Min billing | 1 ms | 100 ms | 1 ms |
| VPC | Yes (cold start impact) | Yes | Yes |
| Provisioned concurrency | Yes | Yes (min instances) | Yes (Premium) |
| Edge | Lambda@Edge | Cloud Functions (limited) | Azure Front Door |

### BaaS Components (Serverless Ecosystem)

| Service | Purpose |
|---------|---------|
| DynamoDB | Serverless NoSQL |
| S3 | Object storage |
| API Gateway | HTTP API management |
| SQS/SNS | Messaging |
| Cognito | Auth |
| Aurora Serverless | Serverless SQL (scales to zero) |

---

## 13. Code or Pseudocode

### Lambda Handler (Python) with Cold Start Optimization

```python
# Lazy init - runs once per container (warm reuse)
_db_client = None
def get_db():
    global _db_client
    if _db_client is None:
        _db_client = create_db_connection()  # Reused across invocations
    return _db_client

def lambda_handler(event, context):
    # Handler runs every invocation
    db = get_db()
    result = db.query("SELECT * FROM orders WHERE id = %s", event['order_id'])
    return {'statusCode': 200, 'body': json.dumps(result)}
```

### Provisioned Concurrency (AWS SAM/CloudFormation)

```yaml
Resources:
  MyFunction:
    Type: AWS::Serverless::Function
    Properties:
      Handler: index.handler
      Runtime: python3.9
      ProvisionedConcurrencyConfig:
        ProvisionedConcurrency: 5  # 5 always-warm instances
```

### Keep-Warm Pattern (Scheduled)

```python
# keep_warm.py - runs every 5 min via EventBridge
def lambda_handler(event, context):
    import urllib.request
    urllib.request.urlopen('https://api.example.com/health')
    return {'status': 'ok'}
```

### S3 Trigger + Image Processing

```python
def lambda_handler(event, context):
    for record in event['Records']:
        bucket = record['s3']['bucket']['name']
        key = record['s3']['object']['key']
        
        # Download, resize, upload
        image = download_from_s3(bucket, key)
        resized = resize_image(image, (200, 200))
        upload_to_s3(bucket, f'thumbnails/{key}', resized)
    
    return {'processed': len(event['Records'])}
```

---

## 14. Interview Discussion

### Key Points
1. **Serverless ≠ No servers**: Servers exist; you don't manage them
2. **Cold start**: Biggest latency concern; mitigate with provisioned concurrency or keep-warm
3. **Stateless**: Design for no in-memory state; use external stores
4. **Cost**: Great for variable/sparse load; can be expensive at high steady load
5. **Vendor lock-in**: Consider abstraction layer for portability
6. **Limits**: Timeout, memory, payload—know your platform limits

### Common Questions
- **"When would you use serverless vs containers?"** → Serverless for event-driven, sporadic; containers for long-running, stateful, predictable high load
- **"How do you handle cold starts?"** → Provisioned concurrency, keep-warm, smaller runtimes, edge (Cloudflare Workers)
- **"What about vendor lock-in?"** → Use Serverless Framework, avoid provider-specific features where possible; accept some lock-in for managed benefits
- **"How do you debug serverless?"** → Distributed tracing (X-Ray, Jaeger), structured logging, local emulation (SAM, Serverless Offline)

### Red Flags
- Using serverless for long-running jobs without decomposition
- Ignoring cold start in latency-sensitive APIs
- No idempotency for event-driven functions (retries = duplicates)
- Storing state in function (won't persist across invocations)

---

## Appendix: Serverless Best Practices

### Function Design
- **Single responsibility**: One function, one purpose
- **Stateless**: Never rely on in-memory state between invocations
- **Idempotent**: Design for at-least-once delivery; handle duplicates
- **Small package**: Minimize dependencies; faster cold start
- **Environment variables**: For config; not secrets (use Secrets Manager)

### Error Handling
- **Retry**: Platform retries async invocations; ensure idempotency
- **Dead letter queue**: Capture failed events for manual handling
- **Structured logging**: JSON logs for parsing; include request ID
- **Alerts**: CloudWatch alarms on error rate, duration, throttles

### Cost Optimization Checklist
- Right-size memory (profile first)
- Reduce duration (optimize code, avoid unnecessary work)
- Use SQS batching (up to 10 messages per invocation)
- Provisioned concurrency only where needed (expensive)
- Consider Spot instances for batch workloads (not Lambda; use ECS/Fargate)

### Vendor Abstraction
- **Serverless Framework**: Deploy to AWS, GCP, Azure with same config
- **Terraform**: Infrastructure as code; multi-cloud
- **Avoid**: Provider-specific features (Lambda Destinations, Step Functions) if portability matters
