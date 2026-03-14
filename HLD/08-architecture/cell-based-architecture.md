# Cell-Based Architecture

## 1. Concept Overview

### Definition
**Cell-based architecture** organizes systems into independent, self-contained deployment units called **cells**. Each cell is a complete copy of the application stack that handles a subset of traffic/users independently. Cells are isolated from each other—a failure in one cell does not cascade to others, and each cell can be deployed, scaled, and operated independently.

### Purpose
- **Blast radius reduction**: Limit impact of failures to a single cell (e.g., 1% of users instead of 100%)
- **Fault isolation**: Hardware/software failures contained within cell boundaries
- **Independent deployment**: Deploy changes to one cell at a time; validate before rolling out
- **Traffic partitioning**: Route users to cells based on consistent hashing, user ID, or geography

### Problems It Solves
- **Cascading failures**: Single bug or misconfiguration taking down entire system
- **Risky deployments**: Need to validate in production with limited exposure
- **Geographic affinity**: Keep user data and traffic in specific regions
- **Regulatory compliance**: Isolate data by region (GDPR, data residency)

---

## 2. Real-World Motivation

### AWS
- **Availability Zones (AZs)**: Each AZ is effectively a cell—independent power, cooling, networking
- **Regional cells**: us-east-1, eu-west-1, etc.—full stack per region
- **Blast radius**: Outage in one AZ doesn't take down the region; regional outage doesn't affect other regions

### Slack
- **Cell architecture**: Users assigned to cells based on workspace ID
- **Independent scaling**: Scale cells based on workspace activity
- **Incident containment**: Bug in message delivery limited to affected cell
- **Deployment**: Canary by cell—deploy to low-traffic cell first

### DoorDash
- **Geographic cells**: Traffic partitioned by region (US, Canada, Australia)
- **Independent stacks**: Each cell has its own databases, caches, services
- **Disaster recovery**: Cell failure triggers failover to backup cell
- **Data locality**: Orders and drivers stay within cell for latency

### Netflix
- **Regional deployment cells**: Deploy to one region, validate, then expand
- **Chaos Engineering**: Chaos Monkey terminates instances within cell; validates isolation

### Microsoft (Azure)
- **Stamp/cell model**: Each stamp is a cell with full application stack
- **Ring deployment**: Inner rings (canary) → outer rings (production)

---

## 3. Architecture Diagrams

### Cell Topology

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        CELL-BASED ARCHITECTURE TOPOLOGY                           │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                   │
│   GLOBAL ROUTING LAYER (Anycast / GeoDNS / Load Balancer)                         │
│   ┌─────────────────────────────────────────────────────────────────────────┐   │
│   │  Route by: user_id hash, geo, or explicit cell assignment                 │   │
│   └─────────────────────────────────────────────────────────────────────────┘   │
│                                      │                                            │
│              ┌───────────────────────┼───────────────────────┐                    │
│              │                       │                       │                    │
│              ▼                       ▼                       ▼                    │
│   ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐           │
│   │     CELL A      │     │     CELL B      │     │     CELL C      │           │
│   │  (Users 0-33%)  │     │  (Users 34-66%) │     │  (Users 67-99%) │           │
│   │                 │     │                 │     │                 │           │
│   │  ┌───────────┐  │     │  ┌───────────┐  │     │  ┌───────────┐  │           │
│   │  │ API GW    │  │     │  │ API GW    │  │     │  │ API GW    │  │           │
│   │  └─────┬─────┘  │     │  └─────┬─────┘  │     │  └─────┬─────┘  │           │
│   │        │        │     │        │        │     │        │        │           │
│   │  ┌─────▼─────┐  │     │  ┌─────▼─────┐  │     │  ┌─────▼─────┐  │           │
│   │  │ Services  │  │     │  │ Services  │  │     │  │ Services  │  │           │
│   │  │ DB, Cache │  │     │  │ DB, Cache │  │     │  │ DB, Cache │  │           │
│   │  └───────────┘  │     │  └───────────┘  │     │  └───────────┘  │           │
│   └─────────────────┘     └─────────────────┘     └─────────────────┘           │
│                                                                                   │
│   Each cell: FULL STACK, ISOLATED, NO SHARED STATE (or minimal cross-cell sync)   │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Routing Layer

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     CELL ROUTING LAYER                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Request arrives with user_id / session_id / tenant_id                   │
│                    │                                                     │
│                    ▼                                                     │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  cell_id = consistent_hash(user_id) % num_cells                   │   │
│   │  OR: cell_id = lookup(user_id)  # explicit assignment table      │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                    │                                                     │
│                    ▼                                                     │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  Route to Cell A / B / C endpoint                                 │   │
│   │  (Different hostnames, IPs, or path prefixes)                     │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│   User "sticky" to cell: same user always goes to same cell              │
│   (enables local caching, session affinity, data locality)               │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Cell Failure Isolation

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     CELL FAILURE ISOLATION                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   NORMAL:  Cell A ──OK──  Cell B ──OK──  Cell C ──OK──                   │
│                                                                          │
│   CELL B FAILS (bug, OOM, DB crash, network partition):                  │
│                                                                          │
│   Cell A ──OK──  Cell B ──FAIL──  Cell C ──OK──                         │
│                    │                                                     │
│                    │  Blast radius: ONLY users in Cell B affected        │
│                    │  Cells A and C: UNAFFECTED                           │
│                    │                                                     │
│   Mitigation options:                                                    │
│   1. Failover: Route Cell B users to backup cell (if data replicated)   │
│   2. Degraded: Show "maintenance" for Cell B users only                  │
│   3. Rollback: Revert Cell B deployment; A and C keep new version        │
│                                                                          │
│   Without cells: Single failure = 100% outage                            │
│   With cells:    Single failure = 1/N outage (N = number of cells)      │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Cell Assignment Strategies

**Consistent Hashing**
- `cell_id = hash(user_id) % num_cells`
- Pros: Deterministic, no lookup table
- Cons: Rebalancing when adding/removing cells requires migration

**Explicit Assignment Table**
- Lookup table: `user_id → cell_id`
- Pros: Fine-grained control, easy migration
- Cons: Extra lookup, table must be highly available

**Geographic**
- Cell = region (us-east, eu-west)
- User routed by IP geolocation
- Pros: Data residency, latency
- Cons: Traveling users may switch cells

**Tenant-Based (B2B)**
- Each customer/workspace assigned to a cell
- Slack: workspace_id → cell
- Pros: Complete isolation for enterprise
- Cons: Uneven load across cells

### Data Isolation
- **Database per cell**: Each cell has its own DB; no cross-cell queries
- **Replication for DR**: Async replicate to backup cell for failover
- **Cross-cell reads**: Avoid when possible; use event-driven sync if needed

### Deployment Flow
1. Deploy to canary cell (lowest traffic)
2. Monitor metrics, errors, latency
3. Gradually roll to more cells
4. Full rollout or rollback per cell

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| Typical cells (large org) | 5-50+ |
| Blast radius (N cells) | 100/N % of users |
| Cell failover time | 30s - 5min (depends on replication) |
| Cross-cell latency | Avoid; 50-200ms if same region |
| Deployment per cell | 5-15 min (independent) |
| User migration (rebalance) | Hours to days (data migration) |

### Slack Scale (Reference)
- Millions of workspaces across cells
- Cell failure: ~2-5% of users affected
- Deployment: Canary cell → 10% → 50% → 100%

### AWS Scale
- 30+ regions, multiple AZs per region
- Each AZ: independent power, networking, compute
- Regional outage: Isolated to that region

---

## 6. Tradeoffs

### Cell-Based vs Traditional Microservices

| Aspect | Traditional Microservices | Cell-Based |
|--------|---------------------------|------------|
| Failure scope | Service-level; can cascade | Cell-level; contained |
| Deployment | All instances together | Per-cell canary |
| Data | Shared DB or per-service | Per-cell DB |
| Routing | Load balancer to any instance | Route to specific cell |
| Complexity | Service mesh, discovery | Cell routing, replication |

### Cell-Based vs Multi-Region

| Aspect | Multi-Region | Cell-Based |
|--------|--------------|------------|
| Granularity | Region (coarse) | Cell (finer) |
| Data | Often replicated globally | Per-cell or async replicate |
| Use case | DR, latency | Blast radius, canary |
| Cost | High (full copy per region) | Medium (cells can share region) |

### Tradeoffs of Cell Architecture

| Benefit | Cost |
|---------|------|
| Blast radius reduction | More infrastructure (N full stacks) |
| Independent deployment | Complex routing, migration |
| Fault isolation | Data replication complexity |
| Canary by cell | Operational overhead (N deployments) |

---

## 7. Variants / Implementations

### Variants

**Shard-Based Cells**
- Cell = shard; users hashed to shard
- Each shard has full stack
- Example: DoorDash, Slack

**Ring-Based (Microsoft)**
- Ring 0: Canary (internal)
- Ring 1: First production
- Ring 2-N: Broader rollout
- Each ring is a cell

**Availability Zone Cells**
- Cell = AZ within region
- Traffic distributed across AZs
- AZ failure = one cell down

### Implementations
- **AWS**: Multi-AZ, regional isolation
- **Slack**: Custom cell routing, workspace assignment
- **DoorDash**: Geographic + shard cells
- **Netflix**: Regional cells, Spinnaker for deployment
- **Kubernetes**: Can use multiple clusters as cells

---

## 8. Scaling Strategies

### Adding Cells
1. Provision new cell (full stack)
2. Update routing: add cell to hash ring or assignment table
3. Migrate users: drain from old cells, assign to new (if rebalancing)
4. Gradual traffic shift

### Scaling Within Cell
- Horizontal scaling of services within each cell
- Each cell scales independently based on its load
- Load balancer within cell (not across cells)

### Cross-Cell Replication
- Async replication for disaster recovery
- Sync for critical cross-cell reads (rare)
- Event-driven: publish events, consume in other cells for eventual consistency

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| Single cell down | Users in that cell affected | Failover to backup cell, or degraded UX |
| Routing layer down | All users affected | Multi-region routing, anycast |
| Bad deployment in one cell | Only that cell | Rollback that cell; others unaffected |
| Data corruption in cell | Cell users | Restore from replica; may lose recent writes |
| Cross-cell dependency failure | Depends on design | Minimize cross-cell deps; circuit breaker |
| Cell rebalancing | Temporary migration load | Gradual migration, off-peak |

### Real Incidents
- **Slack (2020)**: Cell-level isolation limited outage to subset of workspaces
- **AWS us-east-1**: Regional outage; other regions unaffected (cell = region)

---

## 10. Performance Considerations

- **Routing latency**: Cell lookup adds 1-5ms; cache assignment
- **Data locality**: User in cell = local DB, cache; low latency
- **Cross-cell**: Avoid; 50-200ms+ if needed
- **Replication lag**: Async replication 100ms-1s; accept for DR
- **Connection pooling**: Per-cell pools; no cross-cell connection sharing

---

## 11. Use Cases

| Use Case | Why Cell-Based |
|----------|----------------|
| **Large-scale SaaS** (Slack, Notion) | Blast radius, tenant isolation |
| **E-commerce** (DoorDash, Uber) | Geographic cells, independent scaling |
| **Cloud providers** (AWS, Azure) | AZ/region isolation, stamp model |
| **Multi-tenant B2B** | Complete tenant isolation per cell |
| **High-availability systems** | Fault containment, canary deployment |
| **Regulatory compliance** | Data residency per cell (region) |

---

## 12. Comparison Tables

### Architecture Comparison Matrix

| Pattern | Blast Radius | Deployment | Data | Complexity |
|---------|--------------|------------|------|------------|
| **Monolith** | 100% | Single deploy | Shared DB | Low |
| **Microservices** | Service-level | Per-service | Per-service DB | High |
| **Multi-Region** | Region-level | Per-region | Replicated | High |
| **Cell-Based** | 1/N (N cells) | Per-cell | Per-cell | Very High |

### When to Use Cell-Based

| Use Cell-Based | Use Alternatives |
|----------------|------------------|
| Need blast radius < 10% | Small scale, accept full outage |
| Canary by user segment | Canary by % traffic (simpler) |
| Regulatory data isolation | Single region OK |
| 10M+ users, high availability | < 1M users, simpler architecture |

---

## 13. Code or Pseudocode

### Cell Routing (Consistent Hash)

```python
import hashlib

class CellRouter:
    def __init__(self, cells: list[str]):
        self.cells = sorted(cells)
        self.num_cells = len(cells)
    
    def get_cell(self, user_id: str) -> str:
        """Route user to cell based on consistent hash."""
        h = int(hashlib.sha256(user_id.encode()).hexdigest(), 16)
        idx = h % self.num_cells
        return self.cells[idx]

# Usage
router = CellRouter(["cell-a", "cell-b", "cell-c"])
cell = router.get_cell("user_12345")  # Always returns same cell for same user
```

### Cell Assignment Table (Explicit)

```python
class CellAssignmentRouter:
    def __init__(self, assignment_store):
        self.store = assignment_store  # Redis, DB, or config
    
    def get_cell(self, user_id: str) -> str:
        cell = self.store.get(f"user:{user_id}:cell")
        if cell:
            return cell
        # Assign new user to least-loaded cell
        cell = self._assign_to_least_loaded()
        self.store.set(f"user:{user_id}:cell", cell)
        return cell
```

### Canary Deployment by Cell

```python
def deploy_to_cells(new_version: str, rollout_plan: list[str]):
    """Rollout: canary cell first, then expand."""
    for cell in rollout_plan:
        deploy_to_cell(cell, new_version)
        if not wait_and_validate(cell, timeout=300):
            rollback_cell(cell)
            raise DeploymentAborted(f"Cell {cell} failed validation")
        # Proceed to next cell
```

---

## 14. Interview Discussion

### Key Points to Articulate
1. **Definition**: Cells are independent, full-stack deployment units; each handles a subset of users
2. **Blast radius**: Failure in one cell affects only that cell's users (1/N)
3. **Routing**: Users consistently routed to same cell (hash or assignment table)
4. **Data**: Each cell has its own DB; replication for DR
5. **Deployment**: Canary by cell—deploy to one, validate, expand
6. **Tradeoff**: More infrastructure and complexity for fault isolation

### Common Interview Questions
- **"How would you limit blast radius of a deployment?"** → Cell-based architecture; deploy to canary cell first
- **"How does Slack handle millions of workspaces?"** → Cell architecture; workspace assigned to cell
- **"Compare cell-based vs multi-region"** → Cells can be within region (finer granularity); multi-region is geographic DR
- **"How do you route users to cells?"** → Consistent hashing on user_id, or explicit assignment table
- **"What if a cell goes down?"** → Failover to backup cell (if replicated), or degraded UX for affected users

### Red Flags to Avoid
- Confusing cells with microservices (cells contain multiple services)
- Ignoring data replication and failover
- Not considering user migration when rebalancing cells
- Over-engineering for small scale

### Ideal Answer Structure
1. Define cell-based architecture
2. Explain blast radius reduction with example (e.g., 10 cells = 10% max impact)
3. Describe routing (consistent hash, assignment)
4. Mention real examples (AWS, Slack, DoorDash)
5. Compare with alternatives (microservices, multi-region)
6. Discuss tradeoffs (complexity, cost)
