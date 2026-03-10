# Consistent Hashing: Staff+ Engineer Deep Dive

## 1. Concept Overview

### Definition
**Consistent hashing** is a distributed hashing scheme that minimizes the number of keys that need to be remapped when the hash table is resized (i.e., when nodes are added or removed). Unlike traditional modular hashing, only K/N keys need to be remapped on average, where K is the number of keys and N is the number of nodes.

### Purpose
- **Minimal rehashing**: When adding/removing cache nodes or database shards, only a fraction of keys move
- **Load distribution**: Keys spread evenly across nodes
- **Horizontal scaling**: Add/remove nodes without full redistribution
- **Distributed caches**: Memcached, Redis Cluster, CDN edge servers

### Why It Exists
**Problem with simple modular hashing:**
```
key → hash(key) % N = node
```
When N changes (add/remove node), **almost ALL keys** get remapped to different nodes. This causes:
- Cache stampede (thundering herd) when nodes added/removed
- Massive data movement in distributed storage
- Temporary unavailability during rebalancing
- Load spikes on new nodes

### Problems It Solves
1. **Rehashing storm**: Adding 1 node to 100 shouldn't invalidate 99% of cache
2. **Hot spots**: Uneven distribution with simple hash
3. **Scaling events**: Rolling deployment of new cache nodes
4. **Failure recovery**: Node failure shouldn't cascade

---

## 2. Real-World Motivation

### DynamoDB
- **Partition key** → hash → partition
- **Consistent hashing**: Adding capacity doesn't require full table scan/migration
- **Virtual nodes**: Multiple hash positions per physical partition for balance

### Cassandra
- **Partition key** → Murmur3 hash → token ring (0 to 2^64-1)
- **Vnodes**: 256 tokens per node by default for even distribution
- **Adding node**: Only adjacent token ranges move

### Akamai CDN
- **Content** → hash → edge server
- **Consistent hashing**: Object stays on same server when fleet changes
- **Scale**: 400,000+ servers globally

### Discord
- **Voice/video** → hash → media server
- **Challenge**: Millions of concurrent connections; need minimal disruption on scale

### Memcached / Redis Cluster
- **Key** → hash slot → node
- **Redis**: 16,384 slots; consistent hashing for slot-to-node mapping
- **Resharding**: Only affected slots move

### Netflix
- **Content routing** → consistent hashing for Open Connect
- **Scale**: 17,000+ servers, minimal churn on topology changes

---

## 3. Architecture Diagrams

### Simple Modular Hashing (Problem)

```
SIMPLE HASH: key % N
=====================
N=3 nodes:                    N=4 nodes (add 1):
Keys 0-2 → Node 0             Keys 0-3 → Node 0
Keys 3-5 → Node 1      →      Keys 4-7 → Node 1
Keys 6-8 → Node 2             Keys 8-11 → Node 2
                               Keys 12-15 → Node 3 (NEW)

EVERY key's mapping changes! 75% remapped!
Cache invalidated, DB migration nightmare
```

### Hash Ring Concept

```
CONSISTENT HASH RING (0 to 2^32-1)
==================================
                   0
                   │
            ┌──────┴──────┐
            │             │
        N2  │             │  N1
      (120) │             │ (280)
            │      K1     │
            │    (200)    │
            │             │
            │       N3    │
            │     (350)   │
            └──────┬──────┘
                   │
                 256

Key K1 hashes to 200 → clockwise → Node N3 (first node >= 200)
```

### Hash Ring with Nodes and Keys

```
                   0
                   │
         N1 ───────┼─────── N4
        (50)       │       (300)
                   │
              K2   │   K1
             (120) │  (200)
                   │
         N2 ───────┼─────── N3
       (150)       │      (250)
                   │
              K3   │
             (350) │
                   │
    K1 → hash(200) → clockwise → N3
    K2 → hash(120) → clockwise → N2
    K3 → hash(350) → clockwise → N1 (wrap around)
```

### Adding a Node (Minimal Replication)

```
BEFORE (3 nodes):              AFTER (add N4 at 275):
                   0                        0
         N1 ───────┼─────── N3     N1 ───────┼─────── N4 ── N3
        (50)       │      (250)   (50)       │      (275) (250)
                   │                         │
              K1   │                    K1   │
             (200) │                   (200) │
                   │                         │
         N2 ───────┼                 N2 ─────┼
       (150)       │               (150)     │

Only keys between N4 and previous node (N3) move!
K1 was on N3, now on N4. K2, K3 unchanged.
~1/N keys move (N=4 → 25%)
```

### Virtual Nodes (Vnodes)

```
WITHOUT VNODES (3 nodes):           WITH VNODES (3 nodes, 2 vnodes each):
    Uneven distribution                  More even distribution

N1 ─────┐                         N1a ── N2a ── N3a ── N1b ── N2b ── N3b
        │   Large gaps                    │
N2 ─────┼   = Hot spots                   │  Smaller segments
        │                                 │  = Better balance
N3 ─────┘                                 │
```

---

## 4. Core Mechanics

### Step-by-Step Algorithm

1. **Create ring**: Hash space 0 to 2^32-1 (or 2^64)
2. **Place nodes**: Hash each node (multiple times for vnodes) → positions on ring
3. **Place key**: Hash key → position on ring
4. **Find node**: Walk clockwise from key position; first node encountered is owner
5. **Wrap-around**: If no node found clockwise, wrap to 0 (first node)

### Virtual Nodes (Vnodes) Mechanics

- **Without vnodes**: Each physical node = 1 position on ring
  - Problem: Random positions can create uneven segments (one node gets 40% of ring)
- **With vnodes**: Each physical node = M positions (e.g., 256)
  - Each position: hash("node_id:replica_id") or hash("node_id#vnode_id")
  - Result: 256 small segments per node; more uniform distribution
  - Tradeoff: More memory (256 entries per node), more computation

### Jump Consistent Hash (Alternative)

- **Algorithm**: Google's jump hash (2014)
- **No ring**: Deterministic algorithm, no explicit ring structure
- **Property**: When bucket count goes N→N+1, only 1/(N+1) keys move
- **Limitation**: Only supports adding buckets at end (sequential), not arbitrary add/remove
- **Use case**: Sharding where N can only increase

### Rendezvous (Highest Random Weight) Hashing

- **Algorithm**: For each key, compute hash(key, node) for all nodes; pick max
- **No ring**: Deterministic, no data structure
- **Property**: When node added/removed, only keys that hashed to that node move
- **Tradeoff**: O(N) to find node (must check all nodes); N = number of nodes
- **Use case**: Small N (e.g., < 100 nodes)

---

## 5. Numbers

### Rehashing Comparison

| Scenario | Simple Modulo | Consistent Hashing |
|----------|---------------|---------------------|
| Add 1 node to 10 | 90% keys move | ~9% keys move |
| Add 1 node to 100 | 99% keys move | ~1% keys move |
| Remove 1 node from 10 | 100% keys move | ~10% keys move |
| Remove 1 node from 100 | 100% keys move | ~1% keys move |

### Virtual Node Count Impact

| Vnodes/Node | Distribution Std Dev | Memory (1000 nodes) | Lookup |
|-------------|---------------------|---------------------|--------|
| 1 | High (uneven) | 1000 entries | O(log N) |
| 100 | Medium | 100K entries | O(log N) |
| 256 | Low (Cassandra default) | 256K entries | O(log N) |
| 1000 | Very low | 1M entries | O(log N) |

### Performance Numbers

| Operation | Complexity | Typical Time |
|-----------|------------|--------------|
| Find node for key | O(log N) with tree | < 1 μs |
| Find node (linear) | O(N) | N=1000: ~10 μs |
| Add node | O(K/N) keys move | Depends on K |
| Memory (ring) | O(N × vnodes) | 256 vnodes: ~1MB per 10K nodes |

### Real System Numbers

| System | Hash Space | Vnodes | Keys/s Lookup |
|--------|------------|--------|---------------|
| Cassandra | 2^64 | 256 | Millions |
| Redis Cluster | 16,384 slots | N/A (slot-based) | Millions |
| Memcached | 2^32 (ketama) | 160 per node | Millions |
| DynamoDB | Partition key hash | Implicit | 10K+ WCU/partition |

---

## 6. Tradeoffs

### Consistent Hashing vs Modular Hashing

| Aspect | Modular | Consistent |
|--------|---------|------------|
| **Add/remove node** | O(K) keys move | O(K/N) keys move |
| **Implementation** | Simple | More complex |
| **Distribution** | Uniform (if hash good) | Can be uneven (use vnodes) |
| **Lookup** | O(1) | O(log N) or O(N) |
| **Use case** | Fixed N | Dynamic N |

### With vs Without Virtual Nodes

| Aspect | No Vnodes | With Vnodes |
|--------|-----------|--------------|
| **Distribution** | Can be uneven | More uniform |
| **Hot spot risk** | Higher | Lower |
| **Memory** | O(N) | O(N × vnodes) |
| **Rebalance** | Same | Same |
| **Complexity** | Simpler | Slightly more |

### Jump Hash vs Consistent Hashing

| Aspect | Jump Hash | Consistent Hashing |
|--------|-----------|-------------------|
| **Add/remove** | Add only (sequential) | Arbitrary |
| **Keys moved** | 1/(N+1) | ~1/N |
| **Complexity** | O(log N) | O(log N) or O(N) |
| **Memory** | O(1) | O(N) |
| **Use case** | Sharding, N grows | Caches, arbitrary topology |

---

## 7. Variants / Implementations

### Ketama (Memcached)
- **Hash**: MD5 (deprecated), now often CRC32 or MurmurHash
- **Vnodes**: 160 per physical node (4 hashes × 40 points per hash)
- **Format**: "host:port-0", "host:port-1", ... for vnode labels

### Cassandra
- **Hash**: Murmur3Partitioner (64-bit)
- **Ring**: 0 to 2^63-1
- **Vnodes**: 256 tokens per node (configurable)
- **Replication**: N consecutive nodes on ring (by replication strategy)

### Redis Cluster
- **Slots**: 16,384 slots (0-16383)
- **Mapping**: CRC16(key) % 16384 = slot
- **Node**: Owns ranges of slots

### DynamoDB
- **Partition key** → hash → partition
- **Partition**: 10GB max, 3000 RCU or 1000 WCU
- **Implicit consistent hashing** in partition management

### Jump Consistent Hash (Algorithm)

```python
def jump_hash(key, num_buckets):
    """Google's jump consistent hash - O(log N)"""
    b, j = -1, 0
    while j < num_buckets:
        b = j
        key = ((key * 2862933555777941757) + 3037000493) & 0xFFFFFFFFFFFFFFFF
        j = int((b + 1) * (float(1 << 31) / float((key >> 33) + 1)))
    return b
```

---

## 8. Scaling Strategies

### Adding Nodes
1. **Compute** new node positions (with vnodes)
2. **Identify** keys that now belong to new node (clockwise from new positions)
3. **Migrate** only those keys (async, background)
4. **Update** routing table (can be immediate for reads; writes may lag)

### Removing Nodes
1. **Identify** keys on failed node (clockwise from next node)
2. **Replicate** to next node(s) or redistribute
3. **Update** routing table

### Hot Spot Mitigation
1. **More vnodes**: 256 → 1000 for flatter distribution
2. **Replicate hot keys**: Multiple copies on ring
3. **Secondary indexing**: Split hot key into sub-keys
4. **Consistent hashing with bounded load**: Google's improvement (capped load per node)

### Bounded Load Consistent Hashing
- **Problem**: Randomness can still create 2x load on some nodes
- **Solution**: Cap each node at (1+ε) × average load
- **Algorithm**: Skip overloaded nodes when walking ring
- **Result**: Guaranteed O(1/ε) maximum load imbalance

---

## 9. Failure Scenarios

### Real Production Issues

**Cassandra Hot Spots**
- **Cause**: Low vnode count, certain keys (e.g., empty string) hashed to same node
- **Mitigation**: Increase vnodes, use composite keys

**Memcached Thundering Herd**
- **Cause**: Node failure; all keys on that node miss; stampede to DB
- **Mitigation**: Consistent hashing (only 1/N keys affected), cache warming

**Redis Cluster Resharding**
- **Cause**: Adding nodes; slot migration
- **Mitigation**: Async migration, redirect during migration (MASKED/ASK)

**Akamai Edge Failures**
- **Cause**: Edge server down; objects need new home
- **Mitigation**: Consistent hashing limits affected objects; replication

### Hot Partition (DynamoDB)
- **Cause**: One partition key gets too much traffic
- **Symptom**: Throttling (ProvisionedThroughputExceeded)
- **Mitigation**: Add suffix to partition key (e.g., user_id + random 0-9), write sharding

---

## 10. Performance Considerations

### Lookup Performance
- **Sorted structure**: Tree or sorted array for O(log N) lookup
- **Binary search**: Ring positions sorted; find first >= key_hash
- **Caching**: Hot keys may cache node; avoid lookup

### Memory
- **Ring**: N nodes × vnodes × (hash + node_id) ≈ 24 bytes per entry
- **1000 nodes, 256 vnodes**: ~6 MB
- **100K nodes**: ~600 MB (consider hierarchical or jump hash)

### Network
- **Rebalancing**: Only affected keys move; reduce network during scale
- **Batching**: Migrate keys in batches to avoid overload

---

## 11. Use Cases

| System | Use Case | Implementation |
|--------|----------|-----------------|
| **DynamoDB** | Partition key → partition | Implicit, managed |
| **Cassandra** | Partition key → token | Murmur3, 256 vnodes |
| **Akamai CDN** | Object URL → edge server | Cache routing |
| **Discord** | User → media server | Voice/video routing |
| **Memcached** | Key → cache node | Ketama, 160 vnodes |
| **Redis Cluster** | Key → slot → node | CRC16, 16K slots |
| **YouTube** | Video chunk → storage | Content distribution |
| **Netflix** | Content → Open Connect | CDN routing |

---

## 12. Comparison Tables

### Hashing Algorithms Comparison

| Algorithm | Add Node | Remove Node | Distribution | Complexity |
|-----------|----------|-------------|--------------|------------|
| Modulo | O(K) move | O(K) move | Uniform | O(1) |
| Consistent | O(K/N) move | O(K/N) move | Vnode-dependent | O(log N) |
| Jump Hash | O(K/(N+1)) | N/A (add only) | Uniform | O(log N) |
| Rendezvous | O(K/N) | O(K/N) | Uniform | O(N) |

### System Implementation Summary

| System | Hash Function | Space | Vnodes | Keys Move (add 1 to N) |
|--------|---------------|-------|--------|-------------------------|
| Memcached | MD5/CRC32 | 2^32 | 160 | ~1/N |
| Cassandra | Murmur3 | 2^64 | 256 | ~1/N |
| Redis | CRC16 | 16384 | Slots | Slot migration |
| DynamoDB | Internal | Partitions | Implicit | Partition split |
| Jump Hash | 64-bit mix | N buckets | N/A | 1/(N+1) |

---

## 13. Code or Pseudocode

### Basic Consistent Hashing

```python
import bisect
import hashlib

class ConsistentHash:
    def __init__(self, nodes=None, vnodes=100):
        self.vnodes = vnodes
        self.ring = {}  # hash -> node
        self.sorted_keys = []
        if nodes:
            for node in nodes:
                self.add_node(node)
    
    def _hash(self, key):
        return int(hashlib.md5(key.encode()).hexdigest(), 16) % (2**32)
    
    def add_node(self, node):
        for i in range(self.vnodes):
            h = self._hash(f"{node}:{i}")
            self.ring[h] = node
            bisect.insort(self.sorted_keys, h)
    
    def remove_node(self, node):
        for i in range(self.vnodes):
            h = self._hash(f"{node}:{i}")
            del self.ring[h]
            self.sorted_keys.remove(h)
    
    def get_node(self, key):
        if not self.ring:
            return None
        h = self._hash(key)
        idx = bisect.bisect_right(self.sorted_keys, h)
        if idx == len(self.sorted_keys):
            idx = 0
        return self.ring[self.sorted_keys[idx]]
```

### Jump Consistent Hash

```python
def jump_hash(key, num_buckets):
    """
    Google's jump consistent hash.
    key: 64-bit key
    num_buckets: number of buckets (1 to 2^64)
    Returns: bucket index (0 to num_buckets-1)
    """
    b = -1
    j = 0
    while j < num_buckets:
        b = j
        key = ((key * 2862933555777941757) + 3037000493) & 0xFFFFFFFFFFFFFFFF
        j = int((b + 1) * (float(1 << 31) / float((key >> 33) + 1)))
    return b
```

### Rendezvous Hashing

```python
def rendezvous_hash(key, nodes):
    """Highest random weight hashing"""
    max_weight = -1
    best_node = None
    for node in nodes:
        h = hash((key, node))
        if h > max_weight:
            max_weight = h
            best_node = node
    return best_node
```

### Hot Spot Mitigation (Bounded Load)

```python
def get_node_bounded_load(ring, key, loads, capacity):
    """
    Get node with bounded load - skip overloaded nodes
    capacity = (1 + epsilon) * avg_load
    """
    h = hash(key)
    for _ in range(len(ring)):
        node = ring.get_next_node(h)
        if loads[node] < capacity:
            return node
        h = hash((key, node))  # Try next
    return ring.get_next_node(hash(key))  # Fallback
```

---

## 14. Interview Discussion

### How to Explain Consistent Hashing
1. **Problem**: "Imagine 100 cache nodes. With key % N, adding 1 node remaps 99% of keys - cache stampede"
2. **Solution**: "Consistent hashing - keys on a ring, nodes on ring. Key goes to clockwise next node"
3. **Result**: "Adding 1 node only moves ~1% of keys to that node"
4. **Vnodes**: "Virtual nodes prevent hot spots - each physical node has many positions"

### When Interviewers Expect It
- **System design**: "Design a distributed cache" → consistent hashing
- **Deep dive**: "How does Memcached distribute keys?"
- **Scaling**: "How do you add cache nodes without thundering herd?"
- **Database**: "How does Cassandra partition data?"

### Key Points to Hit
- Solves rehashing storm (only K/N keys move)
- Hash ring: keys and nodes on ring; clockwise = owner
- Virtual nodes for even distribution
- Used by: DynamoDB, Cassandra, Memcached, CDNs
- Tradeoff: Slightly more complex than modulo

### Follow-Up Questions
- "How do you handle hot spots?"
- "What are virtual nodes and why use them?"
- "What's the difference between jump hash and consistent hashing?"
- "How would you implement this?"
- "What happens when a node fails?"

### Common Mistakes
- Saying "no keys move" (wrong - ~1/N move)
- Forgetting vnodes (leads to uneven distribution)
- Not explaining the ring (clockwise walk)
- Confusing with simple hashing (modulo)
