# Company Architecture Deep Dives

---

## 1. How Does Netflix Work?

### The Problem
Netflix streams to 200M+ subscribers globally. Need to serve personalized recommendations, handle device diversity (TV, mobile, web), manage CDN delivery, and ensure high availability. A monolith would be a single point of failure and limit scaling.

### The Architecture/Solution

**Microservices**
- Hundreds of services; each owns a domain (recommendations, playback, billing, etc.)
- Independent deployment and scaling
- Evolved from monolith after 2008 database corruption crisis

**Zuul (API Gateway)**
- Front door for all client requests
- Routing, load balancing, authentication
- Integrates with Eureka for service discovery
- Filters for logging, rate limiting

**Eureka (Service Discovery)**
- Dynamic service registry
- Services register on startup; clients discover via Eureka
- No hardcoded IPs; handles scaling and failures

**Hystrix (Fault Tolerance)**
- Circuit breaker: stop calling failing services
- Fallback: return cached/default when downstream fails
- Prevents cascading failures
- Bulkhead: isolate thread pools per dependency

**CDN**
- Open Connect: Netflix's own CDN appliances in ISPs
- Video segments cached at edge
- Origin only for cache misses

### Key Numbers
| Metric | Value |
|--------|-------|
| Subscribers | 200M+ |
| Microservices | 700+ |
| CDN | Open Connect (ISP edge) |
| Regions | Global |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐
    │  Clients    │
    │  (TV/Mobile)│
    └──────┬──────┘
           │
    ┌──────▼──────┐
    │  Zuul       │
    │  (API GW)   │
    └──────┬──────┘
           │
    ┌──────▼──────┐
    │  Eureka     │
    │  (Discovery)│
    └──────┬──────┘
           │
    ┌──────▼──────────────────────────┐
    │  Microservices                   │
    │  ┌────────┐ ┌────────┐ ┌────────┐
    │  │Recommend│ │Playback│ │Billing │
    │  │Hystrix │ │Hystrix │ │Hystrix │
    │  └────────┘ └────────┘ └────────┘
    └─────────────────────────────────┘
           │
    ┌──────▼──────┐
    │  Open Connect│
    │  (CDN)      │
    └─────────────┘
```

### Key Takeaways
- **Zuul + Eureka + Hystrix** = gateway + discovery + resilience
- **Microservices** enable independent scaling and deployment
- **CDN** is critical for video delivery
- **Circuit breaker** prevents cascade failures

### Interview Tip
"Netflix uses Zuul as API gateway, Eureka for service discovery, and Hystrix for circuit breaking. This pattern—gateway, discovery, fault tolerance—is the standard microservices stack. Zuul routes, Eureka finds services, Hystrix protects against failures."

---

## 2. How Netflix Uses Chaos Engineering

### The Problem
In a distributed system with hundreds of services, failures are inevitable. How do you ensure the system degrades gracefully? Traditional testing doesn't catch unexpected failure modes. Need to proactively inject failures in production.

### The Architecture/Solution

**Chaos Monkey**
- Randomly terminates production instances
- Runs during business hours (Simian Army)
- Forces teams to build for failure
- "If it's not in Chaos Monkey's path, it's not production-ready"

**Fault Injection**
- Kill instances, inject latency, return errors
- Chaos Kong: take down entire AZ
- Latency Monkey: add network delay
- Conformity Monkey: find instances that don't follow best practices

**Game Days**
- Planned chaos exercises
- Entire team participates; observe and fix
- Builds muscle memory for incidents

**Principles**
- Start small; expand blast radius over time
- Automate; run continuously
- Measure impact; have rollback

### Key Numbers
| Metric | Value |
|--------|-------|
| Chaos Monkey | Terminates random instances |
| Frequency | Continuous (business hours) |
| Impact | Improved resilience |

### Architecture Diagram (ASCII)

```
    ┌─────────────────────────────────────────┐
    │  Chaos Engineering Pipeline             │
    │  ┌─────────────┐  ┌─────────────┐       │
    │  │ Chaos       │  │ Latency     │       │
    │  │ Monkey      │  │ Monkey      │       │
    │  └──────┬──────┘  └──────┬──────┘       │
    │         │                │              │
    │         └────────┬───────┘              │
    │                  │                     │
    │         ┌────────▼────────┐            │
    │         │  Production     │            │
    │         │  Services       │            │
    │         │  (Failure!)     │            │
    │         └─────────────────┘            │
    └─────────────────────────────────────────┘
```

### Key Takeaways
- **Proactive failure injection** finds weaknesses before users do
- **Chaos Monkey** is the iconic example; many variants exist
- **Automate** chaos; don't rely on manual game days only
- **Build for failure**; chaos engineering validates it

### Interview Tip
"Netflix's Chaos Monkey randomly kills production instances to force resilience. Chaos engineering = proactively inject failures to validate that the system degrades gracefully. Start with non-critical services, expand over time."

---

## 3. How Google Search Works

### The Problem
Index the entire web (hundreds of billions of pages) and return relevant results in milliseconds. Crawl, parse, rank, and serve at scale. Handle queries that have never been seen before.

### The Architecture/Solution

**Crawling**
- Web crawlers (e.g., Googlebot) follow links
- Distributed crawling; politeness (robots.txt, rate limiting)
- Discover new pages; re-crawl for updates
- Store raw HTML

**Indexing**
- Parse HTML; extract text, links, metadata
- Build inverted index: term → list of (doc_id, position)
- Distributed index; sharded by term
- PageRank and other signals computed during indexing

**Ranking (PageRank + ML)**
- PageRank: links = votes; iterative algorithm
- Hundreds of signals: content, links, user behavior, freshness
- Neural networks (BERT, etc.) for relevance
- Personalized results (location, history)

**Serving**
- Query parsed; terms extracted
- Lookup inverted index; get candidate docs
- Rank candidates; return top N
- Sub-second latency; massive parallelism

### Key Numbers
| Metric | Value |
|--------|-------|
| Indexed pages | Hundreds of billions |
| Latency | < 1 second |
| Crawlers | Distributed |
| Index | Inverted, sharded |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐
    │  Web        │
    └──────┬──────┘
           │
    ┌──────▼──────┐
    │  Crawler    │
    │  (Discover) │
    └──────┬──────┘
           │
    ┌──────▼──────┐
    │  Parser     │
    │  (Extract)  │
    └──────┬──────┘
           │
    ┌──────▼──────┐
    │  Indexer    │
    │  (Inverted) │
    └──────┬──────┘
           │
    ┌──────▼──────┐     ┌─────────────┐
    │  Query      │────►│  User       │
    │  Processor  │     │  (Search)   │
    └──────┬──────┘     └─────────────┘
           │
    ┌──────▼──────┐
    │  Ranker     │
    │  (PageRank, │
    │   ML)       │
    └─────────────┘
```

### Key Takeaways
- **Crawl → Index → Rank → Serve** is the pipeline
- **Inverted index** is the core data structure
- **PageRank** was revolutionary; now one of many signals
- **Distributed** at every stage

### Interview Tip
"Google Search: crawl the web, build inverted index (term → doc list), rank with PageRank + ML, serve in sub-second. The inverted index is the key—lookup by term, not by doc."

---

## 4. How Reddit Works

### The Problem
Reddit serves millions of users with posts, comments, votes, and subreddits. Need to handle high read volume, real-time updates, and complex ranking (hot, new, top). Scale with a small team.

### The Architecture/Solution

**Stack**
- Python (web); Cassandra (posts, comments)
- Memcache for caching
- RabbitMQ/Kafka for async (votes, notifications)
- Elasticsearch for search

**Data Model**
- Posts: subreddit, author, title, body, votes, timestamp
- Comments: tree structure; stored in Cassandra
- Votes: denormalized into post/comment for read performance

**Caching**
- Hot posts cached in Memcache
- Subreddit listings cached
- User sessions in Redis

**Queues**
- Vote processing async (eventual consistency)
- Notification fan-out
- Search index updates

### Key Numbers
| Metric | Value |
|--------|-------|
| Users | 50M+ DAU |
| Posts | Billions |
| DB | Cassandra |
| Cache | Memcache, Redis |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐
    │  Clients    │
    └──────┬──────┘
           │
    ┌──────▼──────┐
    │  Load       │
    │  Balancer   │
    └──────┬──────┘
           │
    ┌──────▼──────────────────────────┐
    │  Python App Servers              │
    └──────┬──────────────────────────┘
           │
    ┌──────┼──────┬──────────────┐
    │      │      │              │
┌───▼──┐ ┌─▼──┐ ┌─▼──┐      ┌───▼───┐
│Memcache│ │Redis│ │Kafka│      │Cassandra│
└───────┘ └────┘ └────┘      └────────┘
```

### Key Takeaways
- **Cassandra** for scalable, flexible schema
- **Caching** for hot content
- **Async** for votes and notifications
- **Python** for rapid iteration

### Interview Tip
"Reddit uses Python, Cassandra for posts/comments, Memcache for hot content, and queues for async processing. Cassandra's flexible schema and horizontal scaling fit Reddit's model. Cache aggressively for hot posts."

---

## 5. Slack Architecture

### The Problem
Slack provides real-time messaging for teams. Need sub-100ms message delivery, presence, typing indicators, search, and file sharing. Scale to millions of concurrent connections.

### The Architecture/Solution

**Evolution**
- Started with PHP; migrated to Hack (PHP with types)
- Real-time: custom WebSocket layer; Erlang for connection handling (at scale)

**Real-Time Messaging**
- WebSocket for persistent connections
- Message fan-out: when user sends, fan out to channel members
- Presence: heartbeat + Redis TTL
- Typing: Pub/Sub

**Database**
- MySQL + Vitess for horizontal scaling
- Sharding by workspace or channel
- Search: Elasticsearch

**File Storage**
- S3 or equivalent for files
- CDN for delivery

### Key Numbers
| Metric | Value |
|--------|-------|
| Concurrent connections | Millions |
| Message latency | < 100ms |
| DB | MySQL + Vitess |
| Real-time | WebSocket, Erlang |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐
    │  Clients    │
    │  (WebSocket)│
    └──────┬──────┘
           │
    ┌──────▼──────┐
    │  Connection │
    │  Manager    │
    │  (Erlang)   │
    └──────┬──────┘
           │
    ┌──────▼──────────────────────────┐
    │  API Layer (Hack/PHP)            │
    └──────┬──────────────────────────┘
           │
    ┌──────┼──────┬──────────────┐
    │      │      │              │
┌───▼──┐ ┌─▼──┐ ┌─▼────┐    ┌───▼───┐
│Redis │ │MySQL│ │Elastic│    │  S3   │
│      │ │Vitess│ │Search│    │(Files)│
└──────┘ └─────┘ └──────┘    └───────┘
```

### Key Takeaways
- **WebSocket** for real-time; Erlang for connection density
- **Vitess** scales MySQL
- **Fan-out** for message delivery
- **Redis** for presence, caching

### Interview Tip
"Slack uses WebSocket for real-time, MySQL + Vitess for persistence, Redis for presence. Erlang handles millions of connections. Message fan-out: when you send, push to all channel members' connections."

---

## 6. How Does Bluesky Work?

### The Problem
Bluesky aims to build decentralized social media. Users should own their data and identity; no single company controls the network. Need interoperability, portability, and censorship resistance.

### The Architecture/Solution

**AT Protocol**
- Authenticated Transfer Protocol
- Open-source; federated network model
- Interoperable apps; users can switch providers

**DID (Decentralized Identifier)**
- DID PLC: self-authenticating identity
- Signing key: manages data; held by PDS
- Recovery key: user-held; can override in 72 hours
- Enables account migration without provider consent

**Personal Data Server (PDS)**
- Hosts user data; manages identity
- User can migrate to different PDS
- Data in signed repositories (Git-like)

**Big Graph Service (BGS)**
- Handles discovery, metrics at scale
- Separate from PDS for scalability

**App Views**
- Independent UIs; read from PDS/BGS
- Bluesky app is one view; others can build different UIs

### Key Numbers
| Metric | Value |
|--------|-------|
| Protocol | AT Protocol |
| Identity | DID PLC |
| Components | PDS, BGS, App Views |
| Migration | User-controlled |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐     ┌─────────────┐
    │  Bluesky    │     │  Other App   │
    │  App        │     │  View       │
    └──────┬──────┘     └──────┬──────┘
           │                   │
           └─────────┬─────────┘
                     │
    ┌────────────────▼────────────────┐
    │  AT Protocol                     │
    │  ┌──────────┐  ┌──────────┐     │
    │  │   PDS    │  │   BGS    │     │
    │  │(User data)│  │(Discovery)│    │
    │  └──────────┘  └──────────┘     │
    │  DID PLC (Identity)              │
    └─────────────────────────────────┘
```

### Key Takeaways
- **DID** enables portable identity
- **PDS** holds user data; user can migrate
- **Federated** model; no single point of control
- **Signed repositories** for data integrity

### Interview Tip
"Bluesky uses the AT Protocol with DID for identity. Users own their data in a PDS; they can migrate. It's federated like email—different providers, interoperable. DID PLC has signing + recovery keys."

---

## 7. WeChat Architecture

### The Problem
WeChat is a "super app" with 1.67B monthly users. Messaging, payments, mini programs, video calls, social feed—all in one app. Need to handle massive scale, low latency, and diverse features.

### The Architecture/Solution

**Super App Model**
- Single app; many services (messaging, pay, mini programs)
- Modular architecture; services independently scaled
- Tencent infrastructure; global deployment

**Messaging**
- Similar to WhatsApp: persistent connections, message queue
- Multi-device sync; offline storage
- End-to-end encryption (optional)

**Mini Programs**
- Lightweight apps within WeChat
- No install; run in sandbox
- JavaScript-based; Tencent runtime

**Payments**
- WeChat Pay integrated
- QR code, NFC
- Links to bank accounts

**Scale**
- Distributed data centers
- Sharding by user_id
- CDN for media

### Key Numbers
| Metric | Value |
|--------|-------|
| Monthly users | 1.67 billion |
| Type | Super app |
| Features | Messaging, pay, mini programs |
| Company | Tencent |

### Architecture Diagram (ASCII)

```
    ┌─────────────────────────────────────────┐
    │  WeChat Client (Super App)               │
    │  ┌────────┐ ┌────────┐ ┌────────┐       │
    │  │Message │ │ Pay    │ │ Mini   │       │
    │  │        │ │        │ │Programs│       │
    │  └────────┘ └────────┘ └────────┘       │
    └────────────────────┬────────────────────┘
                         │
    ┌────────────────────▼────────────────────┐
    │  Tencent Backend                          │
    │  ┌──────────┐ ┌──────────┐ ┌──────────┐  │
    │  │ Messaging│ │ Payment │ │ Mini App │  │
    │  │ Service  │ │ Service │ │ Runtime  │  │
    │  └──────────┘ └──────────┘ └──────────┘  │
    └──────────────────────────────────────────┘
```

### Key Takeaways
- **Super app** = many services in one client
- **Mini programs** = lightweight, no install
- **Modular backend**; each service scales independently
- **1.67B users** requires global, sharded infrastructure

### Interview Tip
"WeChat is a super app: messaging, payments, mini programs in one. Architecture is modular—each service scales independently. Mini programs run in a sandbox, no install. Scale via sharding and global DCs."

---

## 8. How Do Apple AirTags Work?

### The Problem
Track lost items (keys, bags) with a small, battery-powered device. Must work when the item is out of Bluetooth range of the owner. Need privacy (Apple doesn't know locations) and precision when nearby.

### The Architecture/Solution

**Bluetooth**
- AirTag broadcasts secure Bluetooth signals
- Low power; battery lasts ~1 year
- Discoverable by nearby Apple devices

**Find My Network (Crowd-Sourcing)**
- Billions of iPhones, iPads, Macs form a mesh
- When any Apple device detects an AirTag, it relays location to iCloud (anonymously, encrypted)
- Owner sees location on map
- No GPS in AirTag; uses finder's location

**UWB (Ultra Wideband)**
- When owner is nearby, UWB enables "Precision Finding"
- Distance and direction; haptic, visual, audio feedback
- Chip in AirTag + iPhone

**Privacy**
- End-to-end encrypted; Apple can't read locations
- Anti-stalking: AirTag notifies nearby iPhones if it's moving with them (not owner)
- NFC: tap to get contact info if found

### Key Numbers
| Metric | Value |
|--------|-------|
| Find My network | 1B+ Apple devices |
| Battery | ~1 year |
| Precision | UWB when nearby |
| Privacy | E2E encrypted |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐
    │  AirTag     │
    │  (Bluetooth)│
    └──────┬──────┘
           │
           │  Broadcast (encrypted)
           │
    ┌──────▼──────────────────────────┐
    │  Find My Network                 │
    │  (Crowd-sourced)                 │
    │  iPhone A ──► iCloud ◄── iPhone B│
    │  (relay)      (encrypted) (relay) │
    └──────┬──────────────────────────┘
           │
    ┌──────▼──────┐
    │  Owner      │
    │  (Find My)  │
    └─────────────┘

    Nearby: UWB → Precision Finding
```

### Key Takeaways
- **Crowd-sourced** location; no GPS in AirTag
- **Privacy**: E2E encrypted; Apple can't see
- **UWB** for precision when owner is close
- **Anti-stalking** features for safety

### Interview Tip
"AirTags use Bluetooth to broadcast. Any nearby Apple device relays location to iCloud (encrypted). Owner sees it on map. No GPS in AirTag—relies on finder's device. UWB for precision when nearby. Crowd-sourced = key insight."

---

## 9. How Apple Pay Handles 41M Transactions/Day

### The Problem
Apple Pay processes 41 million transactions daily. Must be secure (no raw card numbers), fast (NFC tap < 1 sec), and work offline (transit). Support multiple devices (iPhone, Watch) and banks globally.

### The Architecture/Solution

**Tokenization**
- Card number never stored or transmitted
- Device-specific token (Device Account Number, DAN)
- Token stored in Secure Element (hardware)
- Token mapped to real PAN only by payment network

**Secure Element**
- Dedicated chip in device
- Isolated from main CPU; tamper-resistant
- Stores tokens, keys
- Required for NFC payment

**NFC**
- Tap to pay; < 1 second
- Communication between device and terminal
- Token + cryptogram sent; not card number

**Flow**
1. User adds card; bank provisions token to Secure Element
2. At checkout: tap → Secure Element generates cryptogram with token
3. Terminal sends to payment network
4. Network resolves token → PAN; routes to bank
5. Auth; response back

**Transit**
- Express Transit: works when device locked/low battery
- Offline mode for subway gates

### Key Numbers
| Metric | Value |
|--------|-------|
| Transactions/day | 41 million |
| Latency | < 1 sec (NFC) |
| Security | Tokenization, Secure Element |
| Offline | Transit mode |

### Architecture Diagram (ASCII)

```
    ┌─────────────┐     ┌─────────────┐
    │  iPhone     │     │  Terminal   │
    │  (NFC)      │◄───►│  (POS)      │
    └──────┬──────┘     └──────┬──────┘
           │                   │
           │ Secure Element    │
           │ Token + Cryptogram│
           │                   │
           └─────────┬─────────┘
                     │
            ┌────────▼────────┐
            │  Payment        │
            │  Network        │
            │  (Token→PAN)    │
            └────────┬────────┘
                     │
            ┌────────▼────────┐
            │  Bank           │
            │  (Auth)         │
            └─────────────────┘
```

### Key Takeaways
- **Tokenization** = no raw card data on device or in transit
- **Secure Element** = hardware isolation for tokens
- **NFC** = fast, contactless
- **Payment network** resolves token; merchant never sees PAN

### Interview Tip
"Apple Pay uses tokenization: device gets a token, not the real card number. Token lives in Secure Element. At tap, cryptogram + token go to payment network, which resolves to PAN. Merchant never sees card number."
