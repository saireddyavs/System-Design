# Disaster Recovery

> Staff+ Engineer Level | FAANG Interview Deep Dive

---

## 1. Concept Overview

**Disaster Recovery (DR)** is the strategy and process for restoring systems and data after a catastrophic failure—whether from natural disaster, cyberattack, human error, or infrastructure failure. DR planning ensures business continuity when primary systems are unavailable.

### Key Definitions

| Term | Definition |
|------|------------|
| **RPO** | Recovery Point Objective — maximum acceptable data loss (measured in time) |
| **RTO** | Recovery Time Objective — maximum acceptable downtime (time to restore) |
| **MTD** | Maximum Tolerable Downtime — business-defined outage limit |
| **WRT** | Work Recovery Time — time to restore data after systems are up |
| **BCP** | Business Continuity Plan — overall strategy for continuing operations |

### The DR Hierarchy

```
RTO ≥ WRT + (restore time)
RPO determines backup/replication strategy
MTD ≥ RTO (RTO must fit within business tolerance)
```

---

## 2. Real-World Motivation

### Why DR Matters

- **Regulatory compliance**: HIPAA, PCI-DSS, SOC 2 require DR plans
- **Business survival**: 40% of businesses never reopen after disaster (FEMA)
- **Data loss impact**: Average cost of data breach: $4.45M (IBM 2023)
- **Reputation**: Extended outages damage customer trust permanently

### Disaster Types

| Type | Examples | DR Consideration |
|------|----------|-------------------|
| **Natural** | Earthquake, flood, hurricane | Geographic distribution |
| **Technical** | Data center failure, network partition | Redundancy, failover |
| **Human** | Accidental deletion, misconfiguration | Backup, audit |
| **Cyber** | Ransomware, breach | Immutable backups, air-gap |
| **Political** | Region封锁, sanctions | Multi-region, multi-cloud |

### Notable DR Incidents

| Company | Incident | DR Lesson |
|---------|----------|-----------|
| GitLab | 2017 database deletion | Single point of failure in backup |
| Code Spaces | 2014 AWS compromise | No DR, company shut down |
| British Airways | 2017 power failure | Single datacenter dependency |
| AWS | 2011 us-east-1 outage | Multi-AZ isn't multi-region |

---

## 3. Architecture Diagrams

### DR Strategy Spectrum

```
     Backup/Restore    Pilot Light    Warm Standby    Hot Standby/Multi-Site
     (Cheapest)                                              (Most Expensive)
            │                │                │                    │
            ▼                ▼                ▼                    ▼
    ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
    │ Backup only  │  │ Minimal DR   │  │ Scaled-down   │  │ Full active   │
    │ Restore when │  │ resources    │  │ replica       │  │ replica       │
    │ needed       │  │ ready to     │  │ always running│  │ always        │
    │              │  │ scale up     │  │               │  │ serving       │
    └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘
    RTO: Hours-Days   RTO: 1-2 hours   RTO: Minutes     RTO: Seconds
    RPO: Hours-Days   RPO: Minutes      RPO: Seconds     RPO: Zero
```

### Multi-Region Active-Active DR

```
                    ┌─────────────────────────────────────────┐
                    │         Global Traffic Manager           │
                    │    (Route 53 / Cloud DNS / Akamai)       │
                    └─────────────────┬─────────────────────┘
                                      │
              ┌───────────────────────┼───────────────────────┐
              │                       │                       │
              ▼                       ▼                       ▼
    ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
    │   Region A       │     │   Region B       │     │   Region C       │
    │   (Primary)      │     │   (Primary)      │     │   (Primary)      │
    │                  │     │                  │     │                  │
    │  App + DB        │◄───►│  App + DB        │◄───►│  App + DB        │
    │  Full stack     │     │  Full stack     │     │  Full stack     │
    │  Active         │     │  Active         │     │  Active         │
    └─────────────────┘     └─────────────────┘     └─────────────────┘
              │                       │                       │
              └───────────────────────┼───────────────────────┘
                                      │
                          Cross-region replication
                          (sync or async)
```

### Backup and Restore Flow

```
    Production                    Backup Storage              DR Site
    ┌─────────┐                  ┌─────────────┐           ┌─────────┐
    │  Live   │─── Backup ──────►│  S3/GCS/    │           │  Cold   │
    │  Data   │    (scheduled)   │  Glacier    │           │  Standby│
    └─────────┘                  └──────┬──────┘           └────┬────┘
                                         │                       │
                                         │    Restore (on        │
                                         │    disaster)          │
                                         └──────────────────────►
                                         │                       │
                                         │    Bring up services  │
                                         │    Point DNS          │
                                         └──────────────────────►
```

### Replication Topology for DR

```
                    PRIMARY REGION
                    ┌─────────────────┐
                    │  Primary DB      │
                    │  (Read/Write)    │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ Sync Replica     │ │ Async Replica   │ │ Async Replica   │
│ (Same Region)   │ │ (DR Region 1)   │ │ (DR Region 2)   │
│ Low latency     │ │ RPO: seconds    │ │ RPO: seconds    │
└─────────────────┘ └─────────────────┘ └─────────────────┘
```

---

## 4. Core Mechanics

### RPO (Recovery Point Objective)

**Definition**: Maximum acceptable amount of data loss, measured in time.

| RPO | Implication | Strategy |
|-----|-------------|----------|
| 24 hours | Can lose 1 day of data | Daily backups |
| 1 hour | Can lose 1 hour | Hourly backups or async replication |
| 5 minutes | Can lose 5 min | Frequent backups or near-sync replication |
| 0 | No data loss acceptable | Synchronous replication |

### RTO (Recovery Time Objective)

**Definition**: Maximum acceptable time to restore service.

| RTO | Implication | Strategy |
|-----|-------------|----------|
| 1 week | Can be down days | Manual restore from backup |
| 4 hours | Can be down hours | Pilot light, scripted restore |
| 1 hour | Can be down ~1 hour | Warm standby |
| 1 minute | Near-instant | Hot standby, active-active |

### The RPO-RTO Matrix

```
                    Low RPO (minimal data loss)
                              │
                              │  Sync replication
                              │  Hot standby
                              │  $$$$
                              │
    High RTO ─────────────────┼───────────────── Low RTO
    (hours/days)              │                 (seconds)
                              │
                              │  Async replication
                              │  Warm standby
                              │  $$
                              │
                    High RPO (acceptable data loss)
                              │
                              │  Backup/restore
                              │  $
```

---

## 5. Numbers

### Typical RPO/RTO by Tier

| Tier | RPO | RTO | Cost | Example |
|------|-----|-----|------|---------|
| **Tier 0** | 0 | 0 | Highest | Real-time financial trading |
| **Tier 1** | < 1 min | < 1 min | Very high | Core transaction systems |
| **Tier 2** | < 1 hour | < 4 hours | High | Customer-facing apps |
| **Tier 3** | < 24 hours | < 24 hours | Medium | Internal tools |
| **Tier 4** | 24+ hours | 24+ hours | Low | Dev/test environments |

### Backup Strategy Timing

| Backup Type | Frequency | Retention | Restore Time |
|-------------|-----------|-----------|--------------|
| Full | Daily/Weekly | 30-90 days | Hours |
| Incremental | Every 6-24 hours | 7-30 days | Hours (need full + incrementals) |
| Differential | Every 6-24 hours | 7-30 days | Faster than incremental |
| Continuous | Real-time | N/A | Seconds (replication) |

### Replication Latency by Type

| Replication | Latency | RPO | Use Case |
|-------------|---------|-----|----------|
| Synchronous | 1-10ms | 0 | Critical transactions |
| Asynchronous | 100ms-5s | Seconds | Most workloads |
| Batch | Minutes-hours | Hours | Analytics, reporting |

---

## 6. Tradeoffs (Comparison Tables)

### DR Strategy Comparison

| Strategy | RPO | RTO | Cost | Complexity | Use Case |
|----------|-----|-----|------|------------|----------|
| **Backup/Restore** | Hours-Days | Hours-Days | $ | Low | Non-critical |
| **Pilot Light** | Minutes | 1-2 hours | $$ | Medium | Moderate criticality |
| **Warm Standby** | Seconds | Minutes | $$$ | High | Critical apps |
| **Hot Standby** | 0 | Seconds | $$$$ | Very High | Mission-critical |
| **Multi-Site Active** | 0 | 0 | $$$$$ | Highest | Zero downtime |

### Backup Type Comparison

| Type | Size | Speed | Restore | Storage Cost |
|------|------|-------|---------|--------------|
| **Full** | 100% | Slow | Fast (single restore) | High |
| **Incremental** | Small | Fast | Slow (chain restore) | Low |
| **Differential** | Growing | Medium | Medium | Medium |
| **Snapshot** | 100% (delta) | Fast | Fast | Medium |

### Sync vs Async Replication

| Aspect | Synchronous | Asynchronous |
|--------|--------------|--------------|
| **RPO** | 0 | Seconds to minutes |
| **Latency** | +1-10ms per write | No added latency |
| **Availability** | Lower (replica down = primary impacted) | Higher |
| **Cost** | Higher (low-latency link) | Lower |
| **Use case** | Financial, healthcare | Most applications |

---

## 7. Variants/Implementations

### Pilot Light Architecture

```
    PRIMARY (Active)                    DR REGION (Pilot Light)
    ┌─────────────────┐                ┌─────────────────┐
    │ Full stack       │                │ Data replicated │
    │ - App servers    │   ────────►   │ - DB replica    │
    │ - Database       │   replication │ - No app servers│
    │ - Cache         │                │ - Config ready  │
    └─────────────────┘                └────────┬────────┘
                                                │
                                    On disaster: │
                                    - Launch app servers (AMI/container)
                                    - Scale up DB to primary
                                    - Update DNS
```

### Warm Standby Architecture

```
    PRIMARY (Active)                    DR REGION (Warm)
    ┌─────────────────┐                ┌─────────────────┐
    │ Full capacity    │                │ Scaled-down     │
    │ 100% traffic    │   ────────►    │ 10-20% capacity │
    │                 │   replication  │ Always running  │
    └─────────────────┘                └────────┬────────┘
                                                │
                                    On disaster: │
                                    - Scale up to 100%
                                    - Promote DB
                                    - Shift traffic
```

### Data Backup Strategies

#### Full Backup
- Complete copy of all data
- Restore: Single step
- Storage: N × full size (N = retention count)

#### Incremental Backup
- Only changes since last backup (full or incremental)
- Restore: Full + all incrementals in order
- Storage: Minimal growth

#### Differential Backup
- Changes since last full backup
- Restore: Full + latest differential
- Storage: Grows until next full

### 3-2-1 Backup Rule

- **3** copies of data (original + 2 backups)
- **2** different media types (e.g., disk + tape/cloud)
- **1** offsite copy (different geographic location)

---

## 8. Scaling Strategies

### Multi-Region DR Scaling

- **Active-active**: All regions serve traffic, scale all
- **Active-passive**: DR region scales on demand during failover
- **Traffic shifting**: Gradual vs immediate (DNS, anycast)

### Data Scaling for DR

- **Sharding**: Each shard replicated independently
- **Partitioning**: Regional data residency, replicate cross-region
- **Eventual consistency**: Accept temporary inconsistency for scale

---

## 9. Failure Scenarios

### Backup Failure

**Scenario**: Backup job fails silently, discovered during restore.

**Prevention**: Backup verification, restore testing, monitoring/alerting.

### Replication Lag

**Scenario**: Failover to replica with significant lag → data loss.

**Prevention**: Monitor replication lag, failover only when lag acceptable.

### Corrupted Backups

**Scenario**: Ransomware encrypts primary and backup (if same system).

**Prevention**: Immutable backups, air-gapped copies, access control.

### Split-Brain in DR

**Scenario**: Both regions think they're primary after network partition.

**Prevention**: Quorum, fencing, manual failover for DR (slower but safer).

### Restore Takes Longer Than RTO

**Scenario**: Restore from backup takes 8 hours, RTO is 4 hours.

**Prevention**: Regular restore drills, optimize restore process, consider warmer standby.

---

## 10. Performance Considerations

### Backup Performance Impact

- **Full backup**: Can impact production (I/O, CPU)
- **Mitigation**: Backup during low-traffic windows, use replicas for backup
- **Incremental**: Minimal impact, only changed blocks

### Replication Performance

- **Sync replication**: Adds latency to every write
- **Async replication**: No write latency, but RPO > 0
- **Batch replication**: Lowest impact, highest RPO

### Restore Performance

- **Parallel restore**: Use multiple streams
- **Storage tier**: Restore from Glacier can take hours
- **Database**: Restore + replay logs adds time

---

## 11. Use Cases

| Use Case | Recommended Strategy | RPO | RTO |
|----------|----------------------|-----|-----|
| E-commerce checkout | Hot standby, sync replication | 0 | < 1 min |
| User profiles | Warm standby, async replication | < 1 min | < 15 min |
| Analytics/reporting | Backup/restore, daily | 24 hours | 24 hours |
| Static assets | Multi-region active, CDN | 0 | 0 |
| Compliance data | Backup + immutable archive | 0 (backup) | 24 hours |

---

## 12. Comparison Tables

### Cloud Provider DR Services

| Provider | Service | RPO | RTO | Notes |
|----------|---------|-----|-----|-------|
| AWS | RDS Multi-AZ | 0 | < 60s | Single region |
| AWS | RDS Cross-Region | Seconds | Minutes | Manual failover |
| AWS | S3 Cross-Region Replication | Seconds | N/A | Async |
| GCP | Cloud SQL HA | 0 | < 60s | Single region |
| GCP | Spanner | 0 | Seconds | Multi-region |
| Azure | SQL Geo-Replication | Configurable | Minutes | |

### DR Testing Frequency

| Test Type | Frequency | Purpose |
|-----------|-----------|---------|
| Backup verification | Daily | Ensure backups complete |
| Restore drill | Quarterly | Validate restore procedure |
| Full DR failover | Annually | End-to-end validation |
| Tabletop exercise | Annually | Team readiness |

---

## 13. Failover Automation & Runbooks

### Automated Failover Triggers

- Health check failure (consecutive)
- Replication lag exceeds threshold
- Manual trigger (incident response)
- Regional outage detection

### Runbook Structure

```markdown
# DR Failover Runbook: [Service Name]

## Prerequisites
- Access to DNS, load balancer, database
- Communication channel (incident bridge)

## Pre-Failover Checklist
- [ ] Confirm primary is unrecoverable
- [ ] Notify stakeholders
- [ ] Verify DR region health

## Failover Steps
1. Promote database replica
2. Update application configuration
3. Scale up DR resources
4. Update DNS/GTM
5. Verify traffic flow
6. Monitor for issues

## Rollback Procedure
- [Steps to fail back to primary]

## Post-Failover
- Root cause analysis
- Update runbook if needed
```

### DR Decision Tree

```
Disaster detected
    │
    ├─► Can primary recover in < RTO? ─► Wait, monitor
    │
    └─► Primary unrecoverable
            │
            ├─► Data loss acceptable? ─► Restore from backup
            │
            └─► Need minimal data loss ─► Failover to replica
```

---

## 14. Real-World Examples

### Netflix Multi-Region

- Active-active across AWS regions
- Cassandra multi-datacenter replication
- Chaos engineering validates DR
- Can lose entire region, traffic shifts in minutes

### AWS Region Failover

- Each service has region failover strategy
- S3: Cross-region replication optional
- RDS: Manual cross-region failover
- DynamoDB: Global tables for multi-region

### Google Globally Distributed Services

- Spanner: Synchronous replication across regions
- Bigtable: Multi-cluster replication
- Cloud Storage: Dual-region, multi-region options

### GitHub 2018 Incident

- Database failover revealed split-brain risk
- Improved: Multiple replicas, automated failover testing
- Runbooks for every scenario

---

## 15. Code/Pseudocode

### Backup Verification Script

```python
def verify_backup(backup_id: str) -> bool:
    """Verify backup integrity without full restore"""
    # 1. Check backup exists and is complete
    if not backup_complete(backup_id):
        return False
    
    # 2. Verify checksums if available
    if not verify_checksums(backup_id):
        return False
    
    # 3. Spot-check: restore sample data to temp location
    sample_restore = restore_sample(backup_id, size="1GB")
    if not validate_restored_data(sample_restore):
        return False
    
    return True
```

### Failover Decision Logic

```python
def should_failover(primary_region: str, dr_region: str) -> bool:
    # Check primary is truly down (not just network blip)
    if primary_recoverable_within_rto():
        return False
    
    # Check DR region is healthy
    if not dr_region_healthy(dr_region):
        return False
    
    # Check replication lag acceptable
    if replication_lag() > max_acceptable_rpo:
        alert("Failover would exceed RPO")
        # May still failover for critical availability
        return require_manual_approval()
    
    return True
```

### RPO/RTO Monitoring

```python
def check_rpo_compliance():
    """Alert if replication lag exceeds RPO"""
    lag = get_replication_lag_seconds()
    rpo_seconds = RPO_MINUTES * 60
    
    if lag > rpo_seconds:
        alert(f"RPO breach: lag {lag}s exceeds RPO {rpo_seconds}s")

def check_rto_readiness():
    """Verify we can meet RTO if failover needed"""
    estimated_restore_time = estimate_failover_duration()
    if estimated_restore_time > RTO_SECONDS:
        alert(f"RTO at risk: estimated restore {estimated_restore_time}s > RTO {RTO_SECONDS}s")
```

---

## 16. Interview Discussion

### Key Talking Points

1. **RPO vs RTO**: Different constraints, often need different strategies
2. **Cost vs availability**: Linear cost increase for exponential availability improvement
3. **Testing is critical**: DR plans that aren't tested will fail when needed
4. **Automation vs manual**: Automated failover is faster but riskier (split-brain)
5. **3-2-1 rule**: Industry standard for backup strategy

### Common Questions

**Q: How do you determine RPO and RTO for a system?**
- Business input: What can we afford to lose? How long can we be down?
- Regulatory: Some industries mandate specific RPO/RTO
- Technical: What's achievable at what cost?
- Often: Business says "zero", engineering negotiates based on cost

**Q: What's the difference between backup and replication?**
- Backup: Point-in-time copy, restore required, lower cost, higher RTO
- Replication: Continuous copy, failover (no restore), higher cost, lower RTO/RPO

**Q: When would you use sync vs async replication?**
- Sync: When RPO must be 0 (financial transactions, healthcare)
- Async: When small data loss acceptable, want lower latency and higher availability

**Q: How do you test DR without causing an outage?**
- Restore to isolated environment
- Use replica for testing (don't promote)
- Tabletop exercises (walk through runbook)
- Chaos engineering in non-production

**Q: What's pilot light vs warm standby?**
- Pilot light: Minimal always-on (data replicated, no app servers). Scale up on disaster. Cheaper.
- Warm standby: Scaled-down but running replica. Faster failover. More expensive.
