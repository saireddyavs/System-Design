# NoSQL Essentials — Key Patterns & Examples

## Interview-Ready NoSQL Knowledge: MongoDB, Redis, Cassandra, DynamoDB

---

## 1. When to Use NoSQL vs SQL

### Decision Framework

```
                        ┌─────────────┐
                        │ Start Here  │
                        └──────┬──────┘
                               │
                    ┌──────────▼──────────┐
                    │ Need ACID           │
                    │ transactions across │──Yes──▶ SQL (PostgreSQL, MySQL)
                    │ multiple tables?    │
                    └──────────┬──────────┘
                               │ No
                    ┌──────────▼──────────┐
                    │ Schema changes      │
                    │ frequently? Data    │──Yes──▶ Document DB (MongoDB)
                    │ is hierarchical?    │
                    └──────────┬──────────┘
                               │ No
                    ┌──────────▼──────────┐
                    │ Simple key-value    │
                    │ lookups? Need       │──Yes──▶ Key-Value (Redis, DynamoDB)
                    │ sub-millisecond?    │
                    └──────────┬──────────┘
                               │ No
                    ┌──────────▼──────────┐
                    │ Write-heavy?        │
                    │ Time-series?        │──Yes──▶ Wide-Column (Cassandra)
                    │ Multi-region?       │
                    └──────────┬──────────┘
                               │ No
                    ┌──────────▼──────────┐
                    │ Graph traversals?   │
                    │ Social networks?    │──Yes──▶ Graph DB (Neo4j)
                    │ Recommendations?    │
                    └──────────┬──────────┘
                               │ No
                    ┌──────────▼──────────┐
                    │ Full-text search?   │──Yes──▶ Elasticsearch
                    └──────────┬──────────┘
                               │ No
                               ▼
                    Start with SQL ── safest default
```

---

## 2. MongoDB — Document Database

### Core Concepts

| Concept | SQL Equivalent | Description |
|---------|---------------|-------------|
| Database | Database | Container for collections |
| Collection | Table | Group of documents |
| Document | Row | JSON-like data unit (BSON) |
| Field | Column | Key-value pair in document |
| `_id` | Primary Key | Auto-generated unique identifier |
| Embedded Document | JOIN + FK | Nested document within document |
| Reference | Foreign Key | ObjectId pointing to another document |

### CRUD Operations

```javascript
// ============================================
// CREATE
// ============================================

// Insert one
db.users.insertOne({
  name: "Alice",
  email: "alice@example.com",
  age: 30,
  address: {
    city: "San Francisco",
    state: "CA",
    zip: "94105"
  },
  tags: ["premium", "early_adopter"],
  createdAt: new Date()
});

// Insert many
db.users.insertMany([
  { name: "Bob", email: "bob@example.com", age: 25, tags: ["basic"] },
  { name: "Carol", email: "carol@example.com", age: 35, tags: ["premium"] }
]);

// ============================================
// READ
// ============================================

// Find all
db.users.find();

// Find with filter
db.users.find({ age: { $gt: 25 } });

// Find one
db.users.findOne({ email: "alice@example.com" });

// Projection (select specific fields)
db.users.find({ age: { $gt: 25 } }, { name: 1, email: 1, _id: 0 });

// Complex queries
db.users.find({
  $and: [
    { age: { $gte: 25 } },
    { tags: "premium" },
    { "address.city": "San Francisco" }  // dot notation for nested
  ]
});

// Sort, limit, skip (pagination)
db.users.find()
  .sort({ age: -1 })     // descending
  .skip(20)               // offset
  .limit(10);             // page size

// Count
db.users.countDocuments({ tags: "premium" });

// Distinct
db.users.distinct("address.city");

// ============================================
// UPDATE
// ============================================

// Update one
db.users.updateOne(
  { email: "alice@example.com" },
  { 
    $set: { age: 31 },                    // set field
    $push: { tags: "verified" },           // add to array
    $currentDate: { updatedAt: true }      // set to current date
  }
);

// Update many
db.users.updateMany(
  { tags: "basic" },
  { $set: { tier: "free" } }
);

// Upsert (insert if not exists)
db.users.updateOne(
  { email: "dave@example.com" },
  { $set: { name: "Dave", age: 28 } },
  { upsert: true }
);

// Array operations
db.users.updateOne(
  { email: "alice@example.com" },
  { 
    $addToSet: { tags: "newsletter" },     // add only if not present
    $pull: { tags: "basic" },              // remove from array
    $inc: { loginCount: 1 }                // increment
  }
);

// ============================================
// DELETE
// ============================================

db.users.deleteOne({ email: "dave@example.com" });
db.users.deleteMany({ status: "deleted" });
```

### Query Operators

```javascript
// Comparison
{ age: { $eq: 30 } }     // equal
{ age: { $ne: 30 } }     // not equal
{ age: { $gt: 25 } }     // greater than
{ age: { $gte: 25 } }    // greater than or equal
{ age: { $lt: 40 } }     // less than
{ age: { $lte: 40 } }    // less than or equal
{ age: { $in: [25, 30, 35] } }        // in array
{ age: { $nin: [25, 30] } }           // not in array

// Logical
{ $and: [{ age: { $gt: 25 } }, { status: "active" }] }
{ $or: [{ age: { $lt: 20 } }, { age: { $gt: 60 } }] }
{ $not: { age: { $gt: 25 } } }
{ $nor: [{ status: "deleted" }, { status: "banned" }] }

// Element
{ phone: { $exists: true } }          // field exists
{ age: { $type: "number" } }          // field type

// Array
{ tags: { $all: ["premium", "verified"] } }   // has all
{ tags: { $size: 3 } }                         // array length
{ tags: { $elemMatch: { $regex: /^pre/ } } }   // element matches

// Regex
{ name: { $regex: /^alice/i } }               // case-insensitive regex
```

### Aggregation Pipeline

The aggregation pipeline is MongoDB's equivalent of SQL's GROUP BY + JOINs + subqueries:

```javascript
// Example: "Top 5 cities by total revenue from premium users"
db.orders.aggregate([
  // Stage 1: Match (WHERE)
  { $match: { 
    status: "completed",
    orderDate: { $gte: ISODate("2024-01-01") }
  }},
  
  // Stage 2: Lookup (JOIN)
  { $lookup: {
    from: "users",
    localField: "userId",
    foreignField: "_id",
    as: "user"
  }},
  { $unwind: "$user" },  // flatten array from lookup
  
  // Stage 3: Match on joined data
  { $match: { "user.tags": "premium" } },
  
  // Stage 4: Group (GROUP BY)
  { $group: {
    _id: "$user.address.city",
    totalRevenue: { $sum: "$total" },
    orderCount: { $sum: 1 },
    avgOrderValue: { $avg: "$total" },
    customers: { $addToSet: "$userId" }
  }},
  
  // Stage 5: Add computed field
  { $addFields: {
    uniqueCustomers: { $size: "$customers" }
  }},
  
  // Stage 6: Sort
  { $sort: { totalRevenue: -1 } },
  
  // Stage 7: Limit
  { $limit: 5 },
  
  // Stage 8: Project (SELECT)
  { $project: {
    city: "$_id",
    totalRevenue: 1,
    orderCount: 1,
    avgOrderValue: { $round: ["$avgOrderValue", 2] },
    uniqueCustomers: 1,
    _id: 0
  }}
]);
```

### Aggregation Pipeline Stages Cheat Sheet

| Stage | SQL Equivalent | Purpose |
|-------|---------------|---------|
| `$match` | WHERE | Filter documents |
| `$group` | GROUP BY | Aggregate |
| `$sort` | ORDER BY | Sort |
| `$limit` | LIMIT | Limit results |
| `$skip` | OFFSET | Skip results |
| `$project` | SELECT | Shape output |
| `$lookup` | JOIN | Join collections |
| `$unwind` | -- | Flatten arrays |
| `$addFields` | AS (alias) | Add computed fields |
| `$count` | COUNT(*) | Count documents |
| `$bucket` | -- | Auto group into ranges |
| `$facet` | -- | Multiple pipelines in parallel |
| `$out` | INSERT INTO ... SELECT | Write results to collection |
| `$merge` | MERGE / UPSERT | Merge results into collection |

### Indexes in MongoDB

```javascript
// Single field index
db.users.createIndex({ email: 1 });           // ascending
db.users.createIndex({ age: -1 });            // descending

// Unique index
db.users.createIndex({ email: 1 }, { unique: true });

// Compound index
db.orders.createIndex({ userId: 1, orderDate: -1 });

// Text index (full-text search)
db.products.createIndex({ name: "text", description: "text" });
db.products.find({ $text: { $search: "wireless bluetooth" } });

// Partial index (like PostgreSQL partial index)
db.orders.createIndex(
  { userId: 1 },
  { partialFilterExpression: { status: "active" } }
);

// TTL index (auto-delete after time)
db.sessions.createIndex({ createdAt: 1 }, { expireAfterSeconds: 3600 });

// Geospatial index
db.locations.createIndex({ coordinates: "2dsphere" });
db.locations.find({
  coordinates: {
    $near: {
      $geometry: { type: "Point", coordinates: [-73.97, 40.77] },
      $maxDistance: 5000  // meters
    }
  }
});
```

### Schema Design: Embedding vs Referencing

```javascript
// ============================================
// EMBEDDING (denormalized) — data in same document
// ============================================
// Use when: 1:1 or 1:few, data always accessed together, atomic updates needed
{
  _id: ObjectId("..."),
  name: "Alice",
  addresses: [                              // embedded
    { label: "home", city: "SF", zip: "94105" },
    { label: "work", city: "SJ", zip: "95110" }
  ],
  orders: [                                 // ⚠️ Only for small arrays!
    { orderId: "ORD-001", total: 99.99, date: ISODate("2024-01-15") }
  ]
}
// Pros: Single read, no joins, atomic
// Cons: 16MB document limit, data duplication, unbounded arrays

// ============================================
// REFERENCING (normalized) — IDs point to other documents
// ============================================
// Use when: 1:many or many:many, large/growing arrays, data shared across documents

// Users collection
{ _id: ObjectId("user1"), name: "Alice" }

// Orders collection (reference user)
{ _id: ObjectId("ord1"), userId: ObjectId("user1"), total: 99.99 }
{ _id: ObjectId("ord2"), userId: ObjectId("user1"), total: 149.99 }

// Pros: No duplication, no size limits
// Cons: Multiple queries or $lookup needed, no atomic cross-document

// ============================================
// HYBRID — embed frequently accessed, reference the rest
// ============================================
{
  _id: ObjectId("..."),
  name: "Alice",
  latestOrder: {                             // embedded summary
    orderId: "ORD-002",
    total: 149.99,
    date: ISODate("2024-01-15")
  },
  orderCount: 42                             // cached count
}
// Full order details in separate collection
```

---

## 3. Redis — Key-Value / In-Memory Data Store

### Core Data Types

| Type | Description | Commands | Use Case |
|------|-------------|----------|----------|
| **String** | Simple key-value | GET, SET, INCR | Caching, counters, sessions |
| **Hash** | Field-value pairs | HGET, HSET, HGETALL | Objects, user profiles |
| **List** | Ordered collection | LPUSH, RPUSH, LRANGE | Queues, recent items |
| **Set** | Unique unordered | SADD, SMEMBERS, SINTER | Tags, unique visitors |
| **Sorted Set** | Unique + scored | ZADD, ZRANGE, ZRANK | Leaderboards, rankings |
| **Stream** | Append-only log | XADD, XREAD, XRANGE | Event streaming |

### Common Patterns

```python
import redis
r = redis.Redis(host='localhost', port=6379, db=0)

# ============================================
# CACHING
# ============================================

# Cache with TTL
r.setex("user:42:profile", 3600, json.dumps({
    "name": "Alice", "email": "alice@example.com"
}))

# Get cached data
profile = r.get("user:42:profile")
if profile:
    return json.loads(profile)
else:
    # Cache miss — fetch from DB and cache
    profile = db.query("SELECT * FROM users WHERE id = 42")
    r.setex("user:42:profile", 3600, json.dumps(profile))
    return profile

# ============================================
# SESSION STORE
# ============================================

# Store session (hash)
r.hset("session:abc123", mapping={
    "user_id": "42",
    "role": "admin",
    "login_time": "2024-01-15T10:30:00"
})
r.expire("session:abc123", 1800)  # 30 min TTL

# Get session
session = r.hgetall("session:abc123")

# ============================================
# RATE LIMITING (sliding window)
# ============================================

def is_rate_limited(user_id, limit=100, window=60):
    key = f"ratelimit:{user_id}"
    now = time.time()
    
    pipe = r.pipeline()
    pipe.zremrangebyscore(key, 0, now - window)  # remove old entries
    pipe.zadd(key, {str(now): now})              # add current request
    pipe.zcard(key)                               # count requests
    pipe.expire(key, window)                      # set TTL
    _, _, count, _ = pipe.execute()
    
    return count > limit

# ============================================
# LEADERBOARD (Sorted Set)
# ============================================

# Add scores
r.zadd("leaderboard:weekly", {"Alice": 2500, "Bob": 1800, "Carol": 3200})

# Get top 10
top_10 = r.zrevrange("leaderboard:weekly", 0, 9, withscores=True)
# [('Carol', 3200), ('Alice', 2500), ('Bob', 1800)]

# Get rank (0-indexed)
rank = r.zrevrank("leaderboard:weekly", "Bob")  # 2

# Increment score
r.zincrby("leaderboard:weekly", 500, "Bob")     # Bob now 2300

# ============================================
# DISTRIBUTED LOCK
# ============================================

def acquire_lock(lock_name, timeout=10):
    identifier = str(uuid.uuid4())
    # SET NX (only if not exists) with expiry
    result = r.set(f"lock:{lock_name}", identifier, nx=True, ex=timeout)
    return identifier if result else None

def release_lock(lock_name, identifier):
    # Lua script for atomic check-and-delete
    lua = """
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
    else
        return 0
    end
    """
    r.eval(lua, 1, f"lock:{lock_name}", identifier)

# ============================================
# PUB/SUB
# ============================================

# Publisher
r.publish("notifications", json.dumps({"user": 42, "type": "new_message"}))

# Subscriber
pubsub = r.pubsub()
pubsub.subscribe("notifications")
for message in pubsub.listen():
    if message["type"] == "message":
        data = json.loads(message["data"])
        process_notification(data)

# ============================================
# QUEUE (List-based)
# ============================================

# Producer: push to queue
r.lpush("job_queue", json.dumps({"task": "send_email", "to": "alice@example.com"}))

# Consumer: blocking pop
job = r.brpop("job_queue", timeout=30)  # blocks up to 30 seconds
if job:
    task = json.loads(job[1])
    process_task(task)
```

---

## 4. Cassandra — Wide-Column Store

### Key Concepts

| Concept | Description |
|---------|-------------|
| **Keyspace** | Database equivalent, defines replication strategy |
| **Table** | Stores data, defined by partition key + clustering columns |
| **Partition Key** | Determines which node stores the data (CRITICAL!) |
| **Clustering Columns** | Sort order within partition |
| **Primary Key** | `(partition_key, clustering_col1, clustering_col2)` |

### CQL (Cassandra Query Language) Examples

```sql
-- ============================================
-- KEYSPACE (Database)
-- ============================================
CREATE KEYSPACE ecommerce
WITH replication = {
    'class': 'NetworkTopologyStrategy',
    'us-east': 3,
    'eu-west': 3
};

USE ecommerce;

-- ============================================
-- TABLE DESIGN: Query-first approach
-- ============================================

-- Rule #1: Design tables for SPECIFIC queries (not entities)
-- Rule #2: Denormalize aggressively (no JOINs!)
-- Rule #3: Partition key must match your WHERE clause

-- Query: "Get all orders for a user, sorted by date"
CREATE TABLE orders_by_user (
    user_id     UUID,
    order_date  TIMESTAMP,
    order_id    UUID,
    total       DECIMAL,
    status      TEXT,
    items       LIST<FROZEN<MAP<TEXT, TEXT>>>,
    PRIMARY KEY ((user_id), order_date, order_id)
) WITH CLUSTERING ORDER BY (order_date DESC, order_id ASC);

-- (user_id)         = partition key  → all orders for a user on same node
-- order_date        = clustering col → sorted within partition
-- order_id          = clustering col → unique within partition

-- Query: "Get all orders by status" (different query = different table!)
CREATE TABLE orders_by_status (
    status      TEXT,
    order_date  TIMESTAMP,
    order_id    UUID,
    user_id     UUID,
    total       DECIMAL,
    PRIMARY KEY ((status), order_date, order_id)
) WITH CLUSTERING ORDER BY (order_date DESC);

-- ============================================
-- CRUD OPERATIONS
-- ============================================

-- INSERT (also acts as UPSERT — no error on duplicate)
INSERT INTO orders_by_user (user_id, order_date, order_id, total, status)
VALUES (uuid(), toTimestamp(now()), uuid(), 99.99, 'pending');

-- SELECT (must include partition key!)
SELECT * FROM orders_by_user WHERE user_id = ?;                    -- ✅ 
SELECT * FROM orders_by_user WHERE user_id = ? AND order_date > ?;  -- ✅ (range on clustering)
-- SELECT * FROM orders_by_user WHERE status = 'pending';           -- ❌ (no partition key!)
-- SELECT * FROM orders_by_user WHERE order_date > ?;               -- ❌ (no partition key!)

-- UPDATE
UPDATE orders_by_user SET status = 'shipped' 
WHERE user_id = ? AND order_date = ? AND order_id = ?;

-- DELETE
DELETE FROM orders_by_user 
WHERE user_id = ? AND order_date = ? AND order_id = ?;

-- TTL (auto-delete after time)
INSERT INTO sessions (session_id, user_id, data)
VALUES (?, ?, ?) USING TTL 3600;  -- expires in 1 hour

-- Lightweight transactions (compare-and-set)
INSERT INTO users (email, name) VALUES ('alice@example.com', 'Alice')
IF NOT EXISTS;  -- only insert if email doesn't exist
```

### Data Modeling Patterns

```
Pattern 1: Time-Series Data
─────────────────────────────
Table: sensor_readings
PK: ((sensor_id, date), timestamp)

Why: Groups one day's readings per partition,
prevents unbounded partition growth.

Pattern 2: Wide Partition for Messaging
─────────────────────────────────────────
Table: messages_by_conversation
PK: ((conversation_id), sent_at, message_id)

Why: All messages in a conversation on same node,
sorted by time.

Pattern 3: Bucketing for Hot Partitions
─────────────────────────────────────────
Table: events_by_user
PK: ((user_id, month_bucket), event_time, event_id)

Why: Prevents single partition from growing too large.
month_bucket = '2024-01' etc.
```

### Anti-Patterns to Avoid

| Anti-Pattern | Problem | Fix |
|-------------|---------|-----|
| Using secondary indexes heavily | Slow, scatter-gather | Create specific table per query |
| Large partitions (>100MB) | Slow reads, compaction issues | Add time bucket to partition key |
| Too many tombstones | Slow reads after deletes | Use TTL instead of explicit deletes |
| Expecting JOINs | No JOIN support | Denormalize into single table |
| Using ALLOW FILTERING | Full table scan | Redesign partition key |

---

## 5. DynamoDB — Managed Key-Value / Document Store

### Core Concepts

| Concept | Description |
|---------|-------------|
| **Table** | Collection of items |
| **Item** | A single record (like a row) |
| **Partition Key (PK)** | Hash key — determines partition |
| **Sort Key (SK)** | Range key — sorts within partition |
| **GSI** (Global Secondary Index) | Alternate PK/SK for different access patterns |
| **LSI** (Local Secondary Index) | Same PK, different SK |

### Single-Table Design Pattern

```
Table: ECommerceTable
====================================================================
PK (Partition Key)    | SK (Sort Key)        | Attributes
====================================================================
USER#alice            | PROFILE              | name, email, age
USER#alice            | ORDER#2024-01-15#001 | total, status, items
USER#alice            | ORDER#2024-01-20#002 | total, status, items
USER#alice            | ADDRESS#home         | street, city, zip
USER#alice            | ADDRESS#work         | street, city, zip
PRODUCT#widget-001    | METADATA             | name, price, stock
PRODUCT#widget-001    | REVIEW#alice         | rating, comment
PRODUCT#widget-001    | REVIEW#bob           | rating, comment
ORDER#2024-01-15#001  | ITEM#widget-001      | quantity, price
ORDER#2024-01-15#001  | ITEM#gadget-002      | quantity, price
====================================================================

GSI1 (for queries by status):
GSI1-PK: status      | GSI1-SK: order_date  | projected attributes
====================================================================
pending               | 2024-01-20           | order_id, user_id, total
shipped               | 2024-01-15           | order_id, user_id, total
====================================================================
```

### Access Patterns

```python
import boto3
dynamodb = boto3.resource('dynamodb')
table = dynamodb.Table('ECommerceTable')

# ============================================
# GET USER PROFILE
# ============================================
response = table.get_item(Key={'PK': 'USER#alice', 'SK': 'PROFILE'})
user = response.get('Item')

# ============================================
# GET ALL ORDERS FOR A USER
# ============================================
response = table.query(
    KeyConditionExpression='PK = :pk AND begins_with(SK, :sk)',
    ExpressionAttributeValues={
        ':pk': 'USER#alice',
        ':sk': 'ORDER#'
    },
    ScanIndexForward=False  # descending (newest first)
)
orders = response['Items']

# ============================================
# GET ORDER WITH ALL ITEMS
# ============================================
response = table.query(
    KeyConditionExpression='PK = :pk',
    ExpressionAttributeValues={
        ':pk': 'ORDER#2024-01-15#001'
    }
)

# ============================================
# CREATE ORDER (transactional)
# ============================================
client = boto3.client('dynamodb')
client.transact_write_items(
    TransactItems=[
        # Add order to user
        {'Put': {'TableName': 'ECommerceTable', 'Item': {
            'PK': {'S': 'USER#alice'},
            'SK': {'S': 'ORDER#2024-01-15#001'},
            'total': {'N': '99.99'},
            'status': {'S': 'pending'}
        }}},
        # Add order items
        {'Put': {'TableName': 'ECommerceTable', 'Item': {
            'PK': {'S': 'ORDER#2024-01-15#001'},
            'SK': {'S': 'ITEM#widget-001'},
            'quantity': {'N': '2'},
            'price': {'N': '49.99'}
        }}},
        # Decrement stock (with condition)
        {'Update': {
            'TableName': 'ECommerceTable',
            'Key': {'PK': {'S': 'PRODUCT#widget-001'}, 'SK': {'S': 'METADATA'}},
            'UpdateExpression': 'SET stock = stock - :qty',
            'ConditionExpression': 'stock >= :qty',
            'ExpressionAttributeValues': {':qty': {'N': '2'}}
        }}
    ]
)

# ============================================
# QUERY GSI (orders by status)
# ============================================
response = table.query(
    IndexName='GSI1',
    KeyConditionExpression='GSI1PK = :status',
    ExpressionAttributeValues={':status': 'pending'},
    ScanIndexForward=False
)
```

---

## 6. NoSQL Comparison Table

| Feature | MongoDB | Redis | Cassandra | DynamoDB |
|---------|---------|-------|-----------|----------|
| **Type** | Document | Key-Value | Wide-Column | Key-Value/Document |
| **Data Model** | JSON documents | Strings, Hashes, Lists, Sets | Rows + Column families | Items with attributes |
| **Schema** | Flexible | Schema-less | Table schema required | Flexible per item |
| **Query Language** | MongoDB Query | Commands | CQL (SQL-like) | PartiQL / API |
| **Joins** | $lookup (limited) | N/A | None | None |
| **Transactions** | Multi-doc (4.0+) | MULTI/EXEC | Lightweight (LWT) | TransactWriteItems |
| **Scaling** | Sharding (manual) | Cluster | Auto-partition | Auto-managed |
| **Consistency** | Configurable | Strong (single) | Tunable | Strong or eventual |
| **Persistence** | Yes (disk) | Optional (RDB/AOF) | Yes (SSTables) | Yes (managed) |
| **Best For** | Content, catalogs | Caching, sessions | Time-series, IoT | Serverless, variable load |
| **Managed** | Atlas | ElastiCache | Astra | DynamoDB (AWS) |
| **Speed** | ms | sub-ms | ms | ms |

---

## 7. NoSQL Interview Questions

### Q1: When would you choose MongoDB over PostgreSQL?
**Answer**: Choose MongoDB when:
- Schema changes frequently (agile development)
- Data is naturally hierarchical/document-shaped (e.g., CMS content, product catalogs)
- You need horizontal scaling without complex sharding logic
- Don't need complex multi-table joins

**But**: PostgreSQL with JSONB can handle many "document" use cases while keeping ACID transactions.

### Q2: How do you handle relationships in MongoDB?
**Answer**:
- **Embedding** (1:few): Put related data inside the document
- **Referencing** (1:many, many:many): Store ObjectId references, use `$lookup` or application-level joins
- **Hybrid**: Embed summary, reference full data

### Q3: Explain Cassandra's partition key design
**Answer**: The partition key determines:
- Which node stores the data (via consistent hashing)
- What you can efficiently query (must be in WHERE clause)

**Design rules**:
- Choose partition key based on your most common query
- Avoid hot partitions (e.g., don't use boolean as partition key)
- Keep partition size < 100MB
- One table per query pattern (denormalize)

### Q4: How does Redis persist data?
**Answer**:
- **RDB** (snapshots): Periodic point-in-time snapshots. Fast recovery, some data loss.
- **AOF** (append-only file): Logs every write. More durable, slower recovery.
- **RDB + AOF**: Both for best durability.
- **No persistence**: Pure cache mode (fastest).

### Q5: DynamoDB single-table design — pros and cons?
**Pros**: Fewer tables, consistent performance, reduced costs, all access patterns from one table.
**Cons**: Complex to design, hard to evolve, difficult to understand for new team members, limited to 25 GSIs.

### Q6: How do you handle a hot partition in DynamoDB?
**Answer**:
- **Write sharding**: Add random suffix to partition key (`ORDER#2024-01-15#shard-3`)
- **GSI overloading**: Use GSIs to redistribute queries
- **DynamoDB Adaptive Capacity**: Automatically rebalances (but has limits)
- **Caching**: DAX (DynamoDB Accelerator) for read-heavy hot keys

---

## 8. Quick Reference: SQL → NoSQL Mapping

| SQL Concept | MongoDB | Redis | Cassandra | DynamoDB |
|------------|---------|-------|-----------|----------|
| SELECT | find() | GET/HGETALL | SELECT | GetItem/Query |
| WHERE | find({filter}) | — | WHERE (limited) | KeyConditionExpression |
| JOIN | $lookup | — | Not supported | Not supported |
| GROUP BY | $group | — | Not supported | Not supported (do in app) |
| ORDER BY | sort() | ZRANGE | CLUSTERING ORDER | ScanIndexForward |
| LIMIT/OFFSET | limit().skip() | LRANGE | LIMIT | Limit |
| INDEX | createIndex() | — | Built into PK design | GSI/LSI |
| TRANSACTION | session.withTransaction() | MULTI/EXEC | IF NOT EXISTS (LWT) | TransactWriteItems |
| COUNT | countDocuments() | DBSIZE/SCARD | COUNT (avoid!) | Select: COUNT |
