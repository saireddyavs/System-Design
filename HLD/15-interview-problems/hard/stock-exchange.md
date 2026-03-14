# Design Stock Exchange

## 1. Problem Statement & Requirements

### Problem Statement
Design a stock exchange system that matches buy and sell orders via a central order book, executes trades with price-time priority, supports multiple order types (market, limit, stop), and distributes market data in real-time with ultra-low latency.

### Functional Requirements
- **Order submission**: Submit market, limit, stop orders; cancel, modify
- **Order matching engine**: Price-time priority; match buy/sell continuously
- **Order book**: Maintain buy/sell sides per symbol; price levels with queue
- **Order types**: Market (immediate), Limit (at price or better), Stop (triggered at price)
- **Trade execution**: Generate trades when orders match; settlement workflow
- **Market data**: Real-time L1 (best bid/ask), L2 (depth), L3 (full order book); ticker, trades
- **Concurrency**: Handle millions of orders/day; sub-millisecond matching

### Non-Functional Requirements
- **Latency**: Order-to-ack < 100μs; match-to-trade < 1ms
- **Throughput**: 1M+ orders/day; peak 10K orders/sec
- **Consistency**: Strong; no double-fill; strict ordering
- **Availability**: 99.99%; financial regulatory compliance
- **Audit**: Immutable order/trade log; replay capability

### Out of Scope
- Clearing and settlement (T+2) — focus on matching
- Regulatory reporting (SEC, FINRA)
- Multi-venue routing (smart order routing)
- Options, futures (equities only)
- Market maker incentives

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Symbols**: 5,000 listed
- **Orders/day**: 10M
- **Peak**: 100x average during market open = 1,200 orders/sec
- **Hot symbols**: AAPL, MSFT, etc. — 50% of volume
- **Trades**: ~60% of orders result in trade
- **Market data subscribers**: 10K (brokers, HFT)

### QPS Calculation
| Operation | Daily Volume | Peak QPS |
|-----------|--------------|----------|
| Order submit | 10M | ~1,200 |
| Order cancel | 2M | ~250 |
| Match events | 6M | ~700 |
| Trade generation | 6M | ~700 |
| Market data (L1) | 10K subs × 5K symbols × 10 upd/s | 500M upd/s (fan-out) |
| Market data (L2) | 10K × 100 symbols × 10 | 10M upd/s |

### Storage (1 year)
- **Orders**: 2.5B × 100B ≈ 250 GB
- **Trades**: 1.5B × 80B ≈ 120 GB
- **Order book snapshots**: 5K × 1KB × 1M ≈ 5 TB (if stored)
- **Audit log**: 5B × 50B ≈ 250 GB

### Latency Budget
- **Order receipt → ack**: 50μs
- **Order → match**: 100μs
- **Match → trade broadcast**: 50μs
- **Total**: < 200μs p99

### Bandwidth
- **Order ingress**: 1.2K × 100B ≈ 120 KB/s
- **Market data out**: 10K × 1KB × 100 upd/s ≈ 1 GB/s (L1+L2)
- **Trade broadcast**: 700 × 80B × 10K ≈ 560 MB/s

---

## 3. API Design

### REST / Binary Protocol (FIX, OUCH)

```
# Order Entry (Binary preferred for latency)
POST   /api/v1/orders
Body (binary): { symbol, side, type, quantity, price?, stop_price?, client_order_id }
Response: { order_id, status: NEW, timestamp }

POST   /api/v1/orders/:id/cancel
Response: { order_id, status: CANCELLED }

GET    /api/v1/orders/:id
GET    /api/v1/orders
Query: symbol, status, from, to

# Market Data (WebSocket / Multicast)
WS /md/l1/:symbol     # Best bid/ask, last trade
WS /md/l2/:symbol     # Top 10 levels
WS /md/trades/:symbol # Trade stream
```

### FIX Protocol (Industry Standard)

```
# New Order (35=D)
8=FIX.4.4|35=D|55=AAPL|54=1|38=100|40=2|44=150.50|11=clord123|...

# Execution Report (35=8)
8=FIX.4.4|35=8|39=0|55=AAPL|150=0|151=0|14=0|6=0|11=clord123|...
```

### WebSocket Market Data

```json
// L1 Update
{ "type": "l1", "symbol": "AAPL", "bid": 150.49, "ask": 150.51, "bid_size": 500, "ask_size": 300, "last": 150.50 }

// Trade
{ "type": "trade", "symbol": "AAPL", "price": 150.50, "size": 100, "timestamp": 1234567890 }

// L2 Update
{ "type": "l2", "symbol": "AAPL", "bids": [[150.49, 500], [150.48, 1000]], "asks": [[150.51, 300], [150.52, 200]] }
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Order book**: In-memory (primary); Redis for persistence
- **Orders, Trades**: PostgreSQL (audit); write-optimized
- **Market data**: Real-time only; no DB for L1/L2
- **Audit log**: Append-only; Kafka + S3 or ClickHouse

### Schema

**Orders (PostgreSQL - Audit)**
```sql
orders (
  order_id BIGSERIAL PRIMARY KEY,
  client_order_id VARCHAR(50),
  symbol VARCHAR(20),
  side VARCHAR(4),           -- BUY, SELL
  type VARCHAR(10),          -- MARKET, LIMIT, STOP
  quantity BIGINT,
  price DECIMAL(10,2),
  stop_price DECIMAL(10,2),
  status VARCHAR(20),        -- NEW, PARTIAL, FILLED, CANCELLED, REJECTED
  filled_quantity BIGINT DEFAULT 0,
  created_at TIMESTAMP(6),
  updated_at TIMESTAMP(6),
  account_id BIGINT
)
```

**Trades**
```sql
trades (
  trade_id BIGSERIAL PRIMARY KEY,
  symbol VARCHAR(20),
  buy_order_id BIGINT,
  sell_order_id BIGINT,
  price DECIMAL(10,2),
  quantity BIGINT,
  created_at TIMESTAMP(6)
)
```

**Order Book (In-Memory Structure)**
```
Per symbol:
  bids: SortedMap<Price, Queue<Order>>  // Descending
  asks: SortedMap<Price, Queue<Order>>  // Ascending
  orders_by_id: Map<OrderId, Order>
```

### Order Book Data Structure (C++/Rust)

```cpp
// Price level: price + queue of orders (FIFO)
struct PriceLevel {
  price: Decimal;
  orders: Queue<Order>;  // FIFO for time priority
};

// Order book
struct OrderBook {
  bids: BTreeMap<Decimal, PriceLevel>;  // Desc
  asks: BTreeMap<Decimal, PriceLevel>;  // Asc
  orders: HashMap<OrderId, OrderRef>;
};
```

---

## 5. High-Level Architecture

### ASCII Architecture Diagram

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │              BROKERS / HFT (Order Entry)                      │
                                    └───────────────────────────┬─────────────────────────────────┘
                                                                  │
                                                                  │ Colocated / Low-latency network
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         ORDER ENTRY GATEWAY                                                      │
│                                    Auth, Rate limit, Validate, Route                                            │
└───────────────────────────┬──────────────────────────────────────────────────────────────────────────────────┘
                            │
                            │ Single-threaded or lock-free queue per symbol
                            ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    MATCHING ENGINE (Core)                                                        │
│                                                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────────────────────┐  │
│  │  Per-Symbol Order Book (In-Memory)                                                                       │  │  │
│  │                                                                                                         │  │  │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐   BIDS (desc)              │  │  │
│  │  │ 150.50  100 │    │ 150.49  500 │    │ 150.48 1000 │    │ 150.47  200 │   150.50 > 150.49 > ...       │  │  │
│  │  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘                              │  │  │
│  │  ─────────────────────────────────────────────────────────────────────────────────────────────────────  │  │  │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐   ASKS (asc)               │  │  │
│  │  │ 150.51  300 │    │ 150.52  200 │    │ 150.53  300 │    │ 150.54  100 │   150.51 < 150.52 < ...       │  │  │
│  │  └─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘                              │  │  │
│  │                                                                                                         │  │  │
│  │  Matching: bid >= best_ask => match; ask <= best_bid => match                                            │  │  │
│  │  Price-time priority: Same price => FIFO                                                                  │  │  │
│  └─────────────────────────────────────────────────────────────────────────────────────────────────────────┘  │  │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                            │
                            │ Trades, Order acks, Book updates
                            ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    OUTBOUND PIPELINE                                                           │
│                                                                                                               │
│  ┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐                             │
│  │ Execution     │   │ Trade         │   │ Market Data   │   │ Audit         │                             │
│  │ Reports       │   │ Broadcast     │   │ Publisher     │   │ Logger        │                             │
│  │ (to brokers)  │   │               │   │ (L1, L2, L3)  │   │ (Kafka/File)  │                             │
│  └───────┬───────┘   └───────┬───────┘   └───────┬───────┘   └───────┬───────┘                             │
│          │                   │                   │                   │                                     │
│          ▼                   ▼                   ▼                   ▼                                     │
│  ┌───────────────┐   ┌───────────────┐   ┌───────────────┐   ┌───────────────┐                             │
│  │ TCP/FIX       │   │ Multicast     │   │ Multicast     │   │ Kafka         │                             │
│  │ to each broker│   │ (trades)      │   │ (L1/L2)       │   │ Append-only   │                             │
│  └───────────────┘   └───────────────┘   └───────────────┘   └───────────────┘                             │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│                         MARKET DATA SUBSCRIBERS (Brokers, HFT, Retail)                          │
└──────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Order Matching Engine (Price-Time Priority)

**Price-Time Priority**
- **Price**: Best price first (highest bid, lowest ask)
- **Time**: Same price → first in, first out (FIFO)

**Matching Algorithm**
1. **Incoming buy** at price P, quantity Q:
   - While Q > 0 and best_ask <= P: Match with best ask
   - Fill: min(Q, ask_qty); generate trade; update both orders
   - If ask fully filled: Remove from book; next best ask
   - If buy fully filled: Done
   - Else: Remaining buy becomes limit order on book
2. **Incoming sell**: Symmetric

**Example**
- Book: Bids 150.50(100), 150.49(500); Asks 150.51(300), 150.52(200)
- Sell 200 @ 150.49 arrives
- Matches: 100 @ 150.50 (fills bid), 100 @ 150.49 (fills part of 150.49 level)
- Trades: (100, 150.50), (100, 150.49)
- Book: Bids 150.49(400); Asks 150.51(300)

### 6.2 Order Types

**Market Order**
- No price; match immediately at best available
- May walk the book (multiple price levels)
- Remaining quantity: Reject or convert to limit at last fill (exchange-dependent)

**Limit Order**
- Price specified; only match at price or better
- Resting: If no match, add to book

**Stop Order (Stop-Loss)**
- Trigger price; when last trade >= trigger (sell) or <= trigger (buy), convert to market
- Not in book until triggered
- Store separately; check on each trade

**Stop-Limit**
- Trigger + limit price; when triggered, becomes limit order

### 6.3 Order Book Management

**Data Structures**
- **Bids**: Sorted map (price → queue); descending (best = highest)
- **Asks**: Sorted map; ascending (best = lowest)
- **Order lookup**: Map order_id → (price_level, queue_position)
- **Cancel**: O(1) with pointer; or mark cancelled, skip on match

**Concurrency**
- **Option A**: Single thread per symbol; lock-free queue for orders
- **Option B**: Partition symbols across threads; no cross-symbol lock
- **Option C**: Lock per symbol; fine-grained
- **Latency**: Avoid locks; prefer single-threaded per symbol

### 6.4 Trade Execution and Settlement

1. **Match**: Engine produces (buy_order_id, sell_order_id, price, qty)
2. **Trade event**: Broadcast to both parties + market data
3. **Execution report**: Send to each broker (fill, partial fill)
4. **Settlement**: T+2; out of scope; clearing house
5. **Audit**: Append trade to immutable log

### 6.5 Market Data Distribution

**L1 (Top of Book)**
- Best bid, best ask, last trade, volume
- Update on every match or book change
- Multicast or WebSocket

**L2 (Depth)**
- Top N levels (e.g., 10) each side
- Update on level change
- Incremental or snapshot + delta

**L3 (Full Book)**
- Every order; for market makers
- High bandwidth; optional

**Multicast**
- UDP multicast for low latency; no ack
- Subscribers in same datacenter/colocation
- Redundant streams (primary + secondary)

### 6.6 Concurrency at Scale

**Per-Symbol Isolation**
- Each symbol has own order book; no shared state
- Orders for AAPL and MSFT can process in parallel
- Partition symbols across cores

**Order Queue**
- Single producer (gateway) → lock-free SPSC queue → consumer (matching engine)
- Batching: Process batch of orders; reduce context switch

**Memory**
- Pre-allocate order objects; object pool
- No malloc in hot path
- Cache-friendly: Order book levels in contiguous memory

### 6.7 Ultra-Low Latency Techniques

- **Kernel bypass**: DPDK, Solarflare; avoid kernel network stack
- **CPU pinning**: Matching thread on dedicated core
- **NUMA**: Order book in local NUMA node
- **Branch prediction**: Avoid unpredictable branches
- **Language**: C++, Rust; no GC
- **Colocation**: Brokers host servers in same datacenter

---

## 7. Scaling

### Horizontal Scaling (Symbol Partitioning)
- **Partition**: Symbol hash → matching engine instance
- **Hot symbols**: AAPL, MSFT on dedicated instances
- **No cross-symbol**: Matching is per-symbol; natural partition

### Order Entry Gateway
- **Stateless**: Validate, route to correct matching engine
- **Load balance**: By symbol; consistent hashing
- **Multiple gateways**: For redundancy

### Market Data
- **Multicast**: One-to-many; no fan-out in server
- **Consolidation**: Single process publishes; multicast
- **Regional**: Replicate to other regions via WAN (higher latency)

### Persistence
- **Async**: Don't block matching on DB write
- **Write-ahead**: Log to local SSD first; Kafka async
- **Recovery**: Replay order log to rebuild book

---

## 8. Failure Handling

### Matching Engine Crash
- **Failover**: Standby replica; sync state via replication
- **Recovery**: Replay order log from last checkpoint
- **State transfer**: Snapshot + incremental log
- **Downtime**: Seconds to minutes

### Network Partition
- **Stop trading**: Halt new orders; prevent inconsistent state
- **Reconnect**: Replay missed messages; reconcile
- **Circuit breaker**: Reject orders if unable to reach engine

### Order Duplicate
- **Idempotency**: client_order_id; reject duplicate
- **At-least-once**: Gateway retries; engine dedupes

### Audit Log Corruption
- **Replication**: Multiple copies; checksum
- **Regulatory**: Must retain; backup to cold storage

### Market Data Lag
- **Separate channel**: Don't block order path
- **Stale data**: Subscribers handle; sequence numbers

---

## 9. Monitoring & Observability

### Key Metrics
- **Latency**: Order-to-ack (μs); match-to-trade (μs)
- **Throughput**: Orders/sec; trades/sec
- **Queue depth**: Pending orders per symbol
- **Spread**: Best ask - best bid
- **Fill rate**: % of orders filled

### Alerts
- Latency p99 > 1ms
- Order rejection rate > 0.01%
- Matching engine crash
- Audit log write failure

### Tracing
- **Order ID**: Trace from entry → match → trade → ack
- **Span**: Each stage; latency breakdown
- **Correlation**: Link execution report to trade

### Audit
- **Immutable**: Append-only; no updates
- **Replay**: Rebuild order book from log
- **Retention**: 7 years (regulatory)

---

## 10. Interview Tips

### Follow-up Questions
- "How would you handle a flash crash (erroneous orders)?"
- "How do you prevent race conditions in the matching engine?"
- "How would you add iceberg (hidden) orders?"
- "How does market data reach subscribers in < 1ms?"
- "How would you implement an auction (open/close)?"

### Common Mistakes
- **Locking in hot path**: Kills latency; use lock-free
- **DB in sync path**: Persist async
- **Ignoring order types**: Stop orders need special handling
- **Single-threaded**: Can scale per-symbol
- **No audit**: Regulatory requirement

### Key Points to Emphasize
- **Price-time priority**: Core matching rule
- **Latency**: Microseconds; kernel bypass, CPU pinning
- **Order book**: Sorted map + FIFO queue per level
- **Market data**: Multicast; low latency
- **Concurrency**: Per-symbol; lock-free queues
- **Audit**: Immutable log; replay

---

## Appendix: Extended Design Details & Walkthrough Scenarios

### A. Matching Algorithm (Pseudocode)

```
function match(order):
  if order.side == BUY:
    while order.quantity > 0 and best_ask <= order.price:
      level = get_best_ask_level()
      fill_qty = min(order.quantity, level.quantity)
      execute_trade(order, level.top_order, fill_qty, level.price)
      order.quantity -= fill_qty
      update_level(level, fill_qty)
    if order.quantity > 0:
      add_to_order_book(order)
  else:  # SELL
    symmetric...
```

### B. Price-Time Priority Example

- **Book**: Bids: 150.50(A:100), 150.50(B:200), 150.49(C:500)
- **Order**: Sell 150 @ 150.50
- **Match**: With A (100) + B (50); A is first at 150.50, then B
- **Trades**: 100 @ 150.50 (A), 50 @ 150.50 (B)
- **Remaining**: B has 150 left at 150.50

### C. Stop Order Handling

- **Stop sell** at trigger 150.00: When last trade <= 150.00, convert to market sell
- **Store**: Pending stops in separate structure; keyed by trigger price
- **On trade**: Check if any stop triggered; move to matching queue
- **Order**: Process triggered stops before new orders (exchange-dependent)

### D. Iceberg (Hidden) Orders

- **Display**: Only show portion (e.g., 100 of 1000)
- **Match**: When displayed portion filled, reveal next 100
- **Book**: One order; multiple "display slices"
- **Complexity**: More state; same matching logic

### E. Opening Auction

- **Pre-market**: Collect orders; no matching
- **Open**: Compute single price that maximizes volume (clearing price)
- **Execute**: All matching orders fill at clearing price
- **Transition**: Switch to continuous matching

### F. Kernel Bypass (DPDK)

- **Problem**: Kernel network stack adds 10-50μs
- **DPDK**: User-space NIC driver; poll mode
- **Benefit**: Sub-10μs packet processing
- **Trade-off**: Dedicated CPU cores; no other work

### G. Lock-Free Queue (SPSC)

- **Producer**: Gateway thread
- **Consumer**: Matching engine thread
- **Implementation**: Ring buffer; atomic head/tail
- **No locks**: Cache line padding to avoid false sharing

### H. Market Data Multicast

- **Group**: One multicast group per symbol (or per feed)
- **Publish**: Matching engine writes to socket; NIC multicasts
- **Subscribe**: Brokers join group; receive all packets
- **Reliability**: UDP; no ack; retransmit via separate TCP if needed
- **Latency**: < 10μs within same rack

### I. Flash Crash Prevention

- **Circuit breaker**: Halt trading if price moves > X% in Y seconds
- **Limit up/down**: Max price move per day
- **Kill switch**: Manual halt
- **Erroneous order**: Cancel; bust trades (exchange decision)
- **Rate limit**: Per broker; prevent runaway algo

### J. Order Book Recovery

1. **Checkpoint**: Snapshot of all books at T
2. **Log**: All orders from T to now
3. **Crash**: Restore snapshot; replay log
4. **Consistency**: Log must be durable before ack
5. **State transfer**: Send snapshot + log to standby

### K. FIX Execution Report (35=8)

```
39=0  (New)
39=1  (Partial fill)
39=2  (Fill)
39=4  (Canceled)
39=8  (Rejected)
150=0 (New)
150=1 (Partial fill)
150=2 (Fill)
14=0  (CumQty)
6=0   (AvgPx)
```

### L. Latency Breakdown (Target)

| Stage | Budget |
|-------|--------|
| Network (colocated) | 5μs |
| Gateway parse | 10μs |
| Queue | 5μs |
| Matching | 20μs |
| Trade gen | 5μs |
| Ack send | 10μs |
| **Total** | **55μs** |

### M. Symbol Partitioning

- **5K symbols**: 64 partitions → ~80 symbols each
- **Hot**: AAPL, MSFT, GOOGL, AMZN, etc. → dedicated partition
- **Hash**: symbol_id % 64
- **Rebalance**: Rare; add partition; migrate symbols

### N. Trade-Through Rule (Historical)

- **Rule**: Must route to best price across venues
- **Single venue**: N/A
- **Multi-venue**: Smart order routing; out of scope

### O. Settlement (T+2) - Simplified

- **T**: Trade date
- **T+2**: Settlement date; cash and shares exchange
- **Clearing house**: Central counterparty; guarantees trade
- **Out of scope**: Matching engine only produces trades; settlement is downstream
