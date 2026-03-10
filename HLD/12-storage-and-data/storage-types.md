# Storage Types: Block, File, Object & Distributed File Systems

## 1. Concept Overview

Storage systems are categorized by their **access model**, **abstraction level**, and **consistency guarantees**. The four primary types are:

1. **Block Storage** — Raw blocks of fixed size. Lowest level. Used by VMs, databases, OS.
2. **File Storage** — Hierarchical namespace (directories, files). POSIX semantics. NFS, CIFS.
3. **Object Storage** — Flat namespace with unique IDs. HTTP-based. Immutable objects. S3, GCS.
4. **Distributed File Systems** — File abstraction over many nodes. HDFS, GFS, Ceph.

Each type optimizes for different workloads: block for low-latency random I/O, file for shared access, object for massive scale and durability, distributed file systems for big data.

---

## 2. Real-World Motivation

- **AWS EBS**: Block storage underpins EC2. Databases (RDS, Aurora) run on EBS for sub-ms latency.
- **Netflix**: Petabytes in S3 for video assets. Object storage's durability (11 nines) and cost are critical.
- **Google**: GFS (Google File System) enabled MapReduce and Bigtable. HDFS is the open-source analog.
- **Ceph**: Used by Red Hat, OpenStack. Single system provides block (RBD), file (CephFS), and object (RGW).
- **NFS in enterprises**: Shared home directories, application binaries. File storage's familiarity and POSIX compliance.

---

## 3. Architecture Diagrams

### 3.1 Block Storage Architecture (AWS EBS)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         EC2 INSTANCE (VM)                                         │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  Operating System (Linux/Windows)                                             │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                           │
│  │  │  /dev/sda1   │  │  /dev/sdb   │  │  /dev/sdc   │  ← Block devices         │
│  │  │  (root vol)  │  │  (data vol) │  │  (log vol)  │                           │
│  │  └──────┬───────┘  └──────┬──────┘  └──────┬──────┘                           │
│  │         │                 │                 │                                │
│  │         └─────────────────┼─────────────────┘                                │
│  │                           │                                                    │
│  │                    Block I/O (read/write sector N)                             │
│  └───────────────────────────┼──────────────────────────────────────────────────┘
└───────────────────────────────┼──────────────────────────────────────────────────┘
                                │
                                │  iSCSI / NVMe over Fabrics (EBS)
                                ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         EBS STORAGE (per-AZ)                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  EBS Volume = replicated blocks within AZ                                      │
│  │  - gp3: 3000 IOPS baseline, 125 MB/s throughput                                │
│  │  - io2: 256,000 IOPS, 4000 MB/s (Block Express)                                │
│  │  - Physical disks behind storage controllers                                    │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.2 File Storage Architecture (NFS / EFS)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    NFS CLIENT (Application Server)                                 │
│  mount -t nfs nfs-server:/export/data /mnt/data                                   │
│  /mnt/data/                                                                       │
│    ├── project_a/                                                                 │
│    │   ├── file1.txt                                                              │
│    │   └── file2.txt                                                              │
│    └── project_b/                                                                 │
│        └── report.pdf                                                             │
│  POSIX: open(), read(), write(), mkdir()                                          │
└─────────────────────────────────────────────────────────────────────────────────┘
                                │
                                │  NFS protocol (TCP 2049)
                                ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    NFS SERVER / EFS                                               │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  Inode table + directory structure                                             │
│  │  - Path → inode mapping                                                        │
│  │  - File metadata (size, permissions, timestamps)                               │
│  │  - Data blocks (on underlying block storage)                                   │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
│  AWS EFS: Distributed across multiple AZs, scale-out                              │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.3 Object Storage Architecture (S3)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    CLIENT (Application)                                            │
│  PUT /bucket/key HTTP/1.1                                                          │
│  GET /bucket/key                                                                   │
│  DELETE /bucket/key                                                                │
│  No directories — flat key namespace (keys can contain "/" as convention)           │
└─────────────────────────────────────────────────────────────────────────────────┘
                                │
                                │  HTTP/REST (HTTPS)
                                ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    S3 API GATEWAY / LOAD BALANCER                                 │
└───────────────────────────────────────────────────────────────────────────────┬─┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    S3 STORAGE LAYER                                               │
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │  Bucket: my-bucket                                                           │ │
│  │  Object: key = "users/123/avatar.png"                                         │ │
│  │  - Object ID (internal)                                                       │ │
│  │  - Metadata (Content-Type, custom key-value)                                  │ │
│  │  - Data (immutable)                                                           │ │
│  │  - Version ID (if versioning enabled)                                          │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
│  Replication: 11 nines durability (replicated across AZs/facilities)               │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.4 Distributed File System (HDFS)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    HDFS CLIENT                                                    │
│  hdfs dfs -put file.txt /data/                                                    │
│  hdfs dfs -cat /data/file.txt                                                     │
└───────────────────────────────────────────────────────────────────────────────┬─┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    NAMENODE (metadata, single point of coordination)               │
│  - Namespace: /data/file.txt → Block IDs [B1, B2, B3]                              │
│  - Block → DataNode mapping (B1: [DN1, DN2, DN3])                                  │
│  - No data flows through NameNode                                                   │
│  - FSNamesystem, BlockManager                                                      │
└───────────────────────────────────────────────────────────────────────────────┬─┘
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│   DataNode 1  │       │   DataNode 2  │       │   DataNode 3  │
│  Block B1     │       │  Block B2     │       │  Block B3     │
│  Block B2     │       │  Block B1     │       │  Block B1     │
│  Block B3     │       │  Block B3     │       │  Block B2     │
│  (replicas)   │       │  (replicas)   │       │  (replicas)   │
└───────────────┘       └───────────────┘       └───────────────┘
        │                       │                       │
        └───────────────────────┴───────────────────────┘
                    Data transfer (block pipeline)
```

### 3.5 Ceph Architecture (CRUSH)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    CEPH CLIENTS                                                   │
│  RBD (block) | CephFS (file) | RGW (object/S3-compatible)                        │
└───────────────────────────────────────────────────────────────────────────────┬─┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    MONITORS (Paxos consensus)                                     │
│  - Cluster map (OSD map, PG map, CRUSH map)                                       │
│  - No data path                                                                   │
└───────────────────────────────────────────────────────────────────────────────┬─┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    CRUSH ALGORITHM                                                 │
│  Object ID → hash → PG (Placement Group) → CRUSH rules → OSD set                   │
│  - Deterministic: no central lookup for placement                                  │
│  - Rule: "replicate across 3 racks"                                                │
└───────────────────────────────────────────────────────────────────────────────┬─┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│  OSD 1        │       │  OSD 2        │       │  OSD 3        │
│  (Object     │       │  (Object     │       │  (Object     │
│   Storage     │       │   Storage     │       │   Storage     │
│   Daemon)     │       │   Daemon)     │       │   Daemon)     │
└───────────────┘       └───────────────┘       └───────────────┘
```

---

## 4. Core Mechanics

### 4.1 Block Storage

- **Unit**: Fixed-size blocks (typically 512 bytes or 4 KB)
- **Addressing**: Logical Block Address (LBA) — block index
- **Operations**: Read block N, Write block N
- **No metadata**: Application/OS manages structure (filesystem on top)
- **Attach model**: Volume attached to one instance at a time (or shared with special protocols like NVMe-oF for multi-attach)

### 4.2 File Storage

- **Unit**: Files and directories
- **Namespace**: Hierarchical (tree)
- **Operations**: open, read, write, seek, mkdir, unlink
- **Metadata**: Inodes (size, permissions, timestamps)
- **Protocols**: NFS (Unix), CIFS/SMB (Windows), POSIX

### 4.3 Object Storage

- **Unit**: Object = (key, value, metadata)
- **Namespace**: Flat (bucket + key)
- **Immutability**: PUT creates new version; no in-place edit
- **Operations**: PUT, GET, DELETE, LIST
- **Metadata**: Key-value (Content-Type, custom)
- **Protocol**: HTTP REST (S3 API)

### 4.4 Distributed File Systems

- **Abstraction**: File (like NFS) but data distributed across nodes
- **Replication**: Blocks replicated (e.g., 3x in HDFS)
- **Metadata**: Centralized (NameNode) or distributed (Ceph MON + CRUSH)
- **Tuning**: Optimized for large sequential reads (e.g., 128 MB blocks in HDFS)

---

## 5. Numbers

| Metric | Value |
|--------|-------|
| S3 durability | 11 nines (99.999999999%) |
| S3 availability | 99.99% (standard) |
| EBS gp3 IOPS | 3,000 baseline, up to 16,000 |
| EBS gp3 throughput | 125 MB/s baseline, up to 1,000 MB/s |
| EBS io2 Block Express IOPS | Up to 256,000 |
| EBS io2 throughput | Up to 4,000 MB/s |
| HDFS block size | 128 MB (default) |
| HDFS replication | 3 (default) |
| GFS chunk size | 64 MB |
| NFS typical latency | 1–10 ms |
| S3 GET latency | 10–100 ms (first byte) |
| EBS latency | 0.5–5 ms |

---

## 6. Tradeoffs (Comparison Table)

| Aspect | Block | File | Object | Distributed FS |
|--------|-------|------|--------|----------------|
| **Access** | Block I/O (LBA) | POSIX (path) | HTTP (key) | POSIX / API |
| **Latency** | Lowest (sub-ms) | Low (ms) | Higher (10–100 ms) | Variable |
| **Scalability** | Per-volume | Per-share | Massive (exabytes) | Cluster |
| **Use case** | DB, VM, OS | Shared files, home dirs | Backup, media, data lake | Big data, analytics |
| **Cost** | High ($/GB) | Medium | Low ($/GB) | Medium (cluster) |
| **Consistency** | Strong | Strong (NFS v4) | Eventually (list) | Varies |
| **Update model** | In-place | In-place | Immutable (versioned) | In-place |

---

## 7. Variants/Implementations

### Block Storage

- **AWS EBS**: gp2, gp3, io1, io2, io2 Block Express
- **Google Persistent Disk**: pd-standard, pd-ssd, pd-balanced
- **Azure Disks**: HDD, SSD, Ultra
- **Ceph RBD**: Open-source block in Ceph

### File Storage

- **NFS**: NFSv3, NFSv4 (with state)
- **CIFS/SMB**: Windows file sharing
- **AWS EFS**: NFS over AWS, multi-AZ
- **GCP Filestore**: NFS for GCE
- **CephFS**: Distributed file in Ceph

### Object Storage

- **AWS S3**: De facto standard
- **Google Cloud Storage**: S3-compatible API
- **Azure Blob**: Block blobs, append blobs
- **MinIO**: S3-compatible, self-hosted

### Distributed File Systems

- **HDFS**: Hadoop ecosystem
- **GFS**: Google (proprietary)
- **Ceph**: RBD + CephFS + RGW
- **GlusterFS**: Scale-out file

### GFS (Google File System) — Seminal Design

The 2003 GFS paper influenced HDFS and modern distributed storage. Key design decisions:

- **Single Master (GFS Master)**: Metadata only; no data path. Clients contact master for chunk locations, then talk directly to chunkservers.
- **Large chunks (64 MB)**: Reduces metadata, enables persistent TCP connections, reduces master load.
- **Append-heavy workload**: Optimized for "append then read" (logs, MapReduce output). No random write.
- **Relaxed consistency**: GFS guarantees "consistent" (all clients see same data) but not "defined" (clients may see padding from failed writes). Application-level checksums handle this.

### HDFS NameNode Federation

For clusters with billions of blocks, a single NameNode becomes a bottleneck. **Federation** introduces multiple NameNodes, each managing a portion of the namespace (e.g., /user, /data). Block pools are partitioned. No cross-NameNode communication for metadata.

### S3 Durability Deep Dive

S3 achieves 11 nines durability through:

1. **Replication**: Objects stored across multiple AZs (Standard); or single AZ (One Zone-IA).
2. **Erasure coding**: For some storage classes, data is split into shards with parity. Can lose multiple shards and reconstruct.
3. **Checksums**: Every request verified; corruption detected and repaired.
4. **Background verification**: Continuous integrity checks.

### Ceph CRUSH Algorithm Detail

CRUSH (Controlled Replication Under Scalable Hashing) maps object IDs to OSDs without a central directory:

1. **Input**: Object ID, replication count, CRUSH map (cluster topology: racks, hosts, OSDs).
2. **Hash**: `hash(object_id) % num_pgs` → Placement Group (PG).
3. **CRUSH**: For each PG, traverse topology (e.g., "choose 1 from each of 3 racks") using deterministic hash. Output: list of OSDs.
4. **No metadata lookup**: Client computes placement; fetches from OSDs directly.

---

## 8. Scaling Strategies

| Type | Scaling Approach |
|------|------------------|
| **Block** | Larger volumes, more volumes, stripe across volumes |
| **File** | EFS auto-scales; NFS scale-out (multiple exports) |
| **Object** | Infinite scale (S3); add buckets for isolation |
| **HDFS** | Add DataNodes; scale NameNode (Federation, HA) |
| **Ceph** | Add OSDs; CRUSH rebalances |

---

## 9. Failure Scenarios

| Failure | Impact | Mitigation |
|---------|--------|------------|
| EBS volume failure | Data loss if single replica | EBS replicated in AZ; snapshot to S3 |
| NameNode failure | HDFS unavailable | HA NameNode (active-standby) |
| DataNode failure | Under-replicated blocks | Re-replicate from other replicas |
| S3 region outage | Object access impaired | Cross-region replication |
| NFS server down | File share unavailable | NFS HA (failover) |
| Ceph OSD failure | Degraded | Replication (e.g., 3x) |

---

## 10. Performance Considerations

- **Block**: Optimize IOPS vs throughput. Use provisioned IOPS (io2) for DB.
- **File**: NFS — small file vs large sequential; caching (client, server).
- **Object**: Multipart upload for large objects; byte-range GET for partial reads.
- **HDFS**: Large blocks reduce metadata; locality (compute near data).

---

## 11. Use Cases

| Type | Use Case |
|------|----------|
| **Block** | PostgreSQL, MySQL, Oracle; EC2 root volume; Kubernetes persistent volumes |
| **File** | Shared home directories; CI artifact storage; legacy apps needing POSIX |
| **Object** | Data lake (S3); backup; static assets (images, video); ML datasets |
| **Distributed FS** | Hadoop/Spark; HBase; analytics workloads |

---

## 12. Comparison Tables

### When to Use Each

| Need | Choose |
|------|--------|
| Lowest latency, random I/O | Block |
| Shared file access, POSIX | File |
| Massive scale, durability, cost | Object |
| Big data, compute-on-data | Distributed FS (HDFS) |

### EBS Volume Types (AWS)

| Type | IOPS | Throughput | Use Case |
|------|------|------------|----------|
| gp2 | 3 IOPS/GB, burst 3000 | 250 MB/s | General |
| gp3 | 3000 baseline, 16K max | 125–1000 MB/s | General, cost-effective |
| io1 | Up to 64,000 | 1000 MB/s | DB, critical |
| io2 | Up to 256,000 | 4000 MB/s | Mission-critical |

---

## 13. Code/Pseudocode

### S3 Object Operations (Python boto3)

```python
import boto3

s3 = boto3.client('s3')

# PUT object
s3.put_object(
    Bucket='my-bucket',
    Key='data/users/123/profile.json',
    Body=json.dumps({'name': 'Alice'}),
    ContentType='application/json',
    Metadata={'user_id': '123'}
)

# GET object
response = s3.get_object(Bucket='my-bucket', Key='data/users/123/profile.json')
data = response['Body'].read()

# Multipart upload (large file)
upload_id = s3.create_multipart_upload(Bucket='my-bucket', Key='large.bin')['UploadId']
# ... upload parts, complete_multipart_upload
```

### HDFS Read (Java)

```java
Configuration conf = new Configuration();
FileSystem fs = FileSystem.get(conf);
Path path = new Path("/data/input/file.txt");

try (FSDataInputStream in = fs.open(path)) {
    byte[] buffer = new byte[4096];
    int bytesRead;
    while ((bytesRead = in.read(buffer)) != -1) {
        // process buffer
    }
}
```

### Block Device Read (Linux)

```c
// Conceptual: read block N from /dev/sdb
int fd = open("/dev/sdb", O_RDONLY);
char block[4096];
lseek(fd, block_number * 4096, SEEK_SET);
read(fd, block, 4096);
```

---

## 14. Interview Discussion

### Key Points

1. **Block is the foundation**: File and object storage are built on block (or similar low-level storage).
2. **Object storage scales infinitely**: No directory hierarchy, flat namespace, HTTP. S3 is the data lake backbone.
3. **11 nines durability**: S3's replication and erasure coding achieve 99.999999999% durability.
4. **HDFS vs object**: HDFS for compute-on-data (Spark, Hive); S3 for data lake + serverless (Athena, EMR).
5. **CRUSH**: Ceph's placement algorithm is deterministic and avoids a central metadata bottleneck.

### Common Questions

**Q: Why can't I run a database on S3?**  
A: S3 is eventually consistent for list operations, has higher latency, and no random write. Databases need low-latency random read/write — block storage.

**Q: When would you use EFS over S3?**  
A: When the application requires POSIX (legacy apps, shared home dirs, NFS mounts). S3 needs application changes for key-based access.

**Q: How does HDFS achieve fault tolerance?**  
A: Replication (default 3x). Each block is on 3 DataNodes. If one fails, the others replicate to a new node. NameNode tracks block locations.

**Q: What is CRUSH?**  
A: Ceph's placement algorithm. Given object ID and cluster map, deterministically computes which OSDs store the object. No central lookup; scales with cluster size.

### S3 Throughput and Request Limits

| Limit | Value |
|-------|-------|
| PUT/COPY/POST/DELETE | 3,500 requests/second per prefix |
| GET/HEAD | 5,500 requests/second per prefix |
| Multipart upload | 100 MB per part (except last) |
| Single PUT | 5 GB max (use multipart for larger) |

### Block Storage Use Case: Database

Databases require:
- **Low latency**: Sub-millisecond for random I/O
- **High IOPS**: Thousands of small reads/writes per second
- **Consistency**: fsync for durability
- **Block-level access**: Bypass filesystem for direct control

This is why RDS, Aurora, and self-managed PostgreSQL run on EBS (block), not EFS (file) or S3 (object).

### Object Storage: Immutability and Versioning

Objects in S3 are immutable: PUT overwrites entirely. For versioning:
- **S3 Versioning**: Each PUT creates new version; old versions retained. Enables "undelete" and point-in-time recovery.
- **Lifecycle policies**: Move old versions to Glacier, or expire after N days.
- **No partial update**: To "update" an object, read, modify, write new version. Or use multipart with part replacement (S3 doesn't support this directly; use versioning or overwrite).

### File Storage: NFS vs CIFS

| Aspect | NFS | CIFS/SMB |
|--------|-----|----------|
| **Origin** | Unix | Windows |
| **Locking** | Advisory (cooperative) | Mandatory |
| **Auth** | IP/host, Kerberos | AD, Kerberos |
| **Use case** | Linux workloads | Windows shares, cross-platform |

### HDFS Write Path

1. Client asks NameNode for block locations
2. NameNode allocates block, returns DataNode list (e.g., DN1, DN2, DN3)
3. Client writes to DN1; DN1 pipelines to DN2; DN2 to DN3
4. When block full, client requests new block; repeat
5. Client closes file; NameNode commits (fsync to edit log)

Read path: Client asks NameNode for block locations; reads directly from DataNodes (prefer local).
