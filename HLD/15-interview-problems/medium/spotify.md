# Design Spotify

## 1. Problem Statement & Requirements

### Problem Statement
Design a music streaming platform like Spotify that enables users to stream music, search for songs/artists/albums, create and share playlists, receive personalized recommendations, and download music for offline listening.

### Functional Requirements
- **Stream music**: Play songs with quality adaptation based on network
- **Search**: Search songs, artists, albums, playlists, lyrics
- **Playlists**: Create, edit, share, collaborate on playlists
- **Recommendations**: Discover Weekly, Daily Mix, radio stations
- **Social features**: Follow artists, share playlists, see what friends listen to
- **Offline download**: Download playlists/albums for offline (encrypted)
- **Multi-device**: Sync across phone, desktop, web, speakers

### Non-Functional Requirements
- **Scale**: 600M+ users, 100M+ songs, 4M+ artists
- **Latency**: Track start < 1 second, search < 200ms
- **Availability**: 99.9%+
- **Quality**: Seamless playback, minimal buffering

### Out of Scope
- Podcasts (separate system)
- Spotify Connect (device handoff)
- Lyrics sync (real-time)
- Live audio (concerts)

---

## 2. Back-of-Envelope Estimation

### Scale Assumptions
- **Users**: 600M total, 250M MAU
- **Songs**: 100M tracks
- **Concurrent streams**: 10% peak = 60M
- **Avg track**: 3.5 min, 3.5 MB (320kbps), 1.2 MB (128kbps typical)

### QPS Calculation
| Operation | Daily Volume | Peak QPS |
|-----------|--------------|----------|
| Stream (chunk requests) | 60M × 2 chunks/min × 60 × 24 ≈ 173B | ~2M |
| Search | 200M | ~2,300 |
| Playlist operations | 100M | ~1,200 |
| Recommendations | 150M | ~1,700 |
| Social (follow, share) | 50M | ~600 |
| Auth/session | 30M logins | ~350 |

### Storage (5 years)
- **Audio files**: 100M × 5MB avg (multiple qualities) ≈ 500 TB
- **Metadata**: 100M tracks × 2KB ≈ 200 GB
- **User data**: 600M × 5KB ≈ 3 TB
- **Playlists**: 1B playlists × 1KB ≈ 1 TB
- **Artwork**: 100M × 200KB ≈ 20 TB

### Bandwidth
- **Peak egress**: 60M × 320 kbps ≈ 19.2 Tbps
- **Typical**: 60M × 128 kbps ≈ 7.7 Tbps

### Cache
- **CDN**: 90%+ of audio from edge
- **Metadata cache**: Redis, 100M tracks × 2KB ≈ 200 GB
- **Recommendation cache**: Pre-computed, 250M × 2KB ≈ 500 GB
- **Search index**: Elasticsearch, 100M docs × 1KB ≈ 100 GB

---

## 3. API Design

### REST Endpoints

```
# Authentication
POST   /api/v1/auth/login
Body: { "email" | "username", "password" } or OAuth
Response: { "access_token", "refresh_token", "expires_in" }

POST   /api/v1/auth/refresh
Body: { "refresh_token" }
Response: { "access_token", "expires_in" }

# Search
GET    /api/v1/search
Query: q=query, type=track|artist|album|playlist, limit=20, offset=0
Response: { "tracks", "artists", "albums", "playlists" }

# Catalog
GET    /api/v1/tracks/:id
GET    /api/v1/artists/:id
GET    /api/v1/artists/:id/albums
GET    /api/v1/albums/:id
GET    /api/v1/albums/:id/tracks

# Playback
POST   /api/v1/me/player/play
Body: { "context_uri", "offset", "position_ms" }

POST   /api/v1/me/player/pause
POST   /api/v1/me/player/next
POST   /api/v1/me/player/previous
PUT    /api/v1/me/player/seek
Query: position_ms=15000

GET    /api/v1/me/player
Response: { "device", "progress_ms", "item", "is_playing" }

GET    /api/v1/audio/stream/:track_id
Query: quality=low|normal|high
Response: Redirect to CDN URL or stream URL

# Playlists
GET    /api/v1/playlists/:id
GET    /api/v1/playlists/:id/tracks
POST   /api/v1/playlists
Body: { "name", "description", "public" }
POST   /api/v1/playlists/:id/tracks
Body: { "uris": ["spotify:track:xxx"], "position" }
DELETE /api/v1/playlists/:id/tracks
Body: { "tracks": [{ "uri", "positions" }] }
PUT    /api/v1/playlists/:id
Body: { "name", "description", "public" }

# Recommendations
GET    /api/v1/recommendations
Query: seed_tracks, seed_artists, seed_genres, limit=20
Response: { "tracks": [...] }

GET    /api/v1/browse/discover-weekly
GET    /api/v1/browse/daily-mix/:id
GET    /api/v1/browse/featured-playlists

# Social
GET    /api/v1/me/following
POST   /api/v1/me/following
Body: { "ids", "type": "artist" | "user" }
GET    /api/v1/users/:id/playlists
GET    /api/v1/playlists/:id/followers

# Offline
POST   /api/v1/me/downloads
Body: { "playlist_id" | "album_id", "quality" }
GET    /api/v1/me/downloads
DELETE /api/v1/me/downloads/:id
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Tracks/Artists/Albums**: Cassandra (read-heavy, partition by id)
- **Users**: MySQL (accounts, billing)
- **Playlists**: Cassandra (partition by playlist_id)
- **Social graph**: Neo4j or Cassandra (follow relationships)
- **Search**: Elasticsearch (inverted index)
- **Recommendations**: Redis + Cassandra (pre-computed)
- **Streaming state**: Redis (active playback)

### Schema

**Tracks (Cassandra)**
```sql
tracks_by_id (
  track_id UUID PRIMARY KEY,
  name VARCHAR(200),
  artist_ids LIST<UUID>,
  album_id UUID,
  duration_ms INT,
  preview_url VARCHAR(500),
  popularity INT,
  audio_features JSON,  -- tempo, key, energy, etc.
  created_at TIMESTAMP
)

tracks_by_artist (
  artist_id UUID,
  track_id UUID,
  PRIMARY KEY (artist_id, track_id)
)
```

**Artists (Cassandra)**
```sql
artists_by_id (
  artist_id UUID PRIMARY KEY,
  name VARCHAR(200),
  genres LIST<VARCHAR>,
  image_url VARCHAR(500),
  popularity INT,
  followers_count BIGINT,
  created_at TIMESTAMP
)
```

**Playlists (Cassandra)**
```sql
playlists_by_id (
  playlist_id UUID PRIMARY KEY,
  owner_id BIGINT,
  name VARCHAR(200),
  description TEXT,
  public BOOLEAN,
  track_count INT,
  image_url VARCHAR(500),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)

playlist_tracks (
  playlist_id UUID,
  position INT,
  track_id UUID,
  added_at TIMESTAMP,
  added_by BIGINT,
  PRIMARY KEY (playlist_id, position)
)
```

**Users (MySQL)**
```sql
users (
  user_id BIGINT PRIMARY KEY,
  email VARCHAR(255) UNIQUE,
  display_name VARCHAR(100),
  password_hash VARCHAR(255),
  plan VARCHAR(20),
  country VARCHAR(2),
  created_at TIMESTAMP
)
```

**Social Graph (Cassandra)**
```sql
follows (
  follower_id BIGINT,
  followee_id BIGINT,
  followee_type VARCHAR(20),  -- artist, user
  created_at TIMESTAMP,
  PRIMARY KEY (follower_id, followee_type, followee_id)
)
```

**Listen History (Cassandra)**
```sql
listen_history (
  user_id BIGINT,
  track_id UUID,
  played_at TIMESTAMP,
  context VARCHAR(50),
  PRIMARY KEY (user_id, played_at)
) WITH CLUSTERING ORDER BY (played_at DESC)
```

---

## 5. High-Level Architecture

### ASCII Architecture Diagram

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                        CLIENTS                               │
                                    │  (Mobile App, Web Player, Desktop, Smart Speakers)            │
                                    └───────────────────────────┬─────────────────────────────────┘
                                                                  │
                                                                  │ HTTPS
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         API GATEWAY (Kong / AWS ALB)                                          │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                                  │
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         MICROSERVICES LAYER                                                    │
│                                                                                                               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │   Auth      │ │   Search    │ │  Catalog    │ │  Playback   │ │  Playlist   │ │ Recommend   │            │
│  │   Service   │ │   Service   │ │  Service   │ │  Service   │ │  Service   │ │   Engine    │            │
│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘            │
│         │              │              │              │              │              │                       │
│         │              │              │              │              │              │                       │
│         ▼              ▼              ▼              ▼              ▼              ▼                       │
│  ┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐    │
│  │                         MESSAGE QUEUE (Kafka)                                                       │    │
│  │  Topics: listen_events, playlist_updates, recommendation_requests                                    │    │
│  └─────────────────────────────────────────────────────────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                                  │
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         DATA LAYER                                                            │
│                                                                                                               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐                         │
│  │ Cassandra│  │  MySQL   │  │  Redis   │  │Elasticsearch│ │  Neo4j   │  │   S3     │                         │
│  │(Tracks,  │  │ (Users)  │  │ (Cache,  │  │ (Search   │  │ (Social  │  │ (Audio   │                         │
│  │Playlists)│  │          │  │ Session) │  │  Index)   │  │  Graph)  │  │  Files)  │                         │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘                         │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
                                                                  │
                                                                  │ Audio Stream URL
                                                                  ▼
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                         CDN (CloudFront / Akamai)                                             │
│                                                                                                               │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐                          │
│  │  Edge Location  │  │  Edge Location  │  │  Edge Location  │  │  Edge Location  │  ...                      │
│  │  (Audio Cache)  │  │  (Audio Cache)  │  │  (Audio Cache)  │  │  (Audio Cache)  │                          │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘                          │
│           │                    │                    │                    │                                    │
│           └────────────────────┴────────────────────┴────────────────────┘                                    │
│                                        │                                                                      │
│                                        ▼                                                                      │
│                              ┌─────────────────┐                                                              │
│                              │  Origin (S3)    │                                                              │
│                              │  Audio Files    │                                                              │
│                              │  (96, 128, 256, │                                                              │
│                              │   320 kbps)     │                                                              │
│                              └─────────────────┘                                                              │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                              AUDIO PROCESSING PIPELINE                                                          │
│                                                                                                               │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐                                                │
│  │  Ingest  │───▶│  Encode  │───▶│  Store   │───▶│ Distribute│                                               │
│  │ (WAV/FLAC│    │(Ogg Vorbis│   │  (S3)    │    │  (CDN)   │                                               │
│  │  Master) │    │ 96-320kbps)│   │          │    │          │                                               │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘                                                │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Component Descriptions
- **Auth Service**: OAuth 2.0, token refresh, device management
- **Search Service**: Elasticsearch over metadata + lyrics
- **Catalog Service**: Track, artist, album metadata
- **Playback Service**: Stream URL generation, quality selection, state
- **Playlist Service**: CRUD, collaboration, sharing
- **Recommendation Engine**: Collaborative filtering, audio analysis, NLP
- **CDN**: Audio file delivery, multiple quality tiers

---

## 6. Detailed Component Design

### 6.1 Audio Streaming

**Chunked Delivery**
- Audio pre-encoded into chunks (e.g., 30-second segments)
- Client requests chunks sequentially; can request next before current ends
- Reduces latency (don't wait for full file)

**Codec Selection**
- **Ogg Vorbis**: Primary codec (open, efficient)
- **AAC**: Fallback for some devices
- **Quality tiers**: 96, 128, 256, 320 kbps

**Quality Adaptation**
- Client detects network (WiFi vs cellular)
- User preference (Data Saver mode)
- Adaptive: Start at 128kbps, upgrade if bandwidth allows
- Pre-buffer: 2-3 chunks ahead

### 6.2 Search Service

**Inverted Index (Elasticsearch)**
- **Tracks**: title, artist name, album name, lyrics
- **Artists**: name, genres
- **Albums**: title, artist
- **Playlists**: name, description
- **Lyrics**: Full-text search (licensed from Musixmatch, etc.)

**Indexing Pipeline**
- New track/artist/album → Kafka → Indexer → Elasticsearch
- Bulk indexing for catalog updates
- Synonym handling: "band" ↔ "artist"

**Ranking**
- Relevance score (TF-IDF, BM25)
- Popularity boost (trending tracks rank higher)
- Personalization (user's listening history)

### 6.3 Recommendation Engine

**Collaborative Filtering**
- User-track matrix: who listened to what
- Matrix factorization (implicit feedback: play count, skip rate)
- "Users like you also listened to X"

**Content-Based (Audio Analysis)**
- **Audio features**: Tempo, key, energy, danceability, valence
- **Similarity**: Cosine similarity on feature vectors
- "Tracks similar to X" (same sonic characteristics)

**NLP on Playlists**
- Playlist names: "Chill Study", "Workout Mix"
- Embed playlists; recommend tracks that fit
- "Add to playlist" suggestions

**Contextual**
- Time of day, day of week, device
- "Morning commute" vs "Evening relax"

**Discover Weekly**
- Batch job: Generate 30 tracks per user weekly
- Blend: Collaborative + content-based + novelty
- Stored in Redis/Cassandra, served via API

### 6.4 Playlist Service

- **CRUD**: Create, add/remove tracks, reorder
- **Collaboration**: Multiple editors (with conflict resolution)
- **Sharing**: Public/private, share link
- **Offline**: Mark playlist for download; encrypt and store locally

### 6.5 Social Graph

- **Follow artists**: Get new release notifications
- **Follow users**: See their public playlists
- **Shared playlists**: Collaborative editing
- **Graph DB (Neo4j)**: For complex queries (friends of friends who like X)
- **Cassandra**: Simpler follow storage for scale

### 6.6 Offline Caching

- **Encrypted storage**: AES-256, key derived from user credentials
- **License**: DRM-like; tracks expire after 30 days offline
- **Sync**: When online, refresh licenses; download new tracks
- **Storage**: Local device (SQLite + encrypted files)

### 6.7 Music Licensing / Rights Management

- **Labels**: Licensing agreements with major labels
- **Royalties**: Per-stream payment; tracked per play
- **Territory**: Different catalog per country (licensing)
- **Metadata**: ISRC, UPC for royalty reporting

---

## 7. Scaling

### Sharding
- **Cassandra**: Partition by track_id, playlist_id, user_id
- **Elasticsearch**: Shards by document ID hash
- **MySQL**: Shard users by user_id range

### Caching
- **CDN**: 90%+ audio from edge
- **Redis**: Catalog metadata, session, playback state
- **Recommendations**: Pre-computed in Redis
- **Cache-aside**: Catalog service

### CDN
- **CloudFront / Akamai**: Global edge locations
- **Pre-positioning**: Popular tracks at edge
- **Origin**: S3 with cross-region replication

### Database
- **Cassandra**: Multi-DC, RF=3
- **Read replicas**: MySQL for read-heavy ops
- **Connection pooling**: Reduce connection overhead

---

## 8. Failure Handling

### Component Failures
- **Search down**: Fallback to "Popular" or cached results
- **Recommendation down**: Show "Trending" playlists
- **Playback**: CDN redundancy; multiple origins
- **Auth**: Token validation cached; refresh can retry

### Redundancy
- **Multi-region**: Active-active in US, EU, APAC
- **Cassandra**: RF=3, QUORUM reads/writes
- **CDN**: Multi-CDN failover
- **Kafka**: Replicated partitions

### Degradation
- **Quality**: Fall back to lower bitrate if CDN slow
- **Search**: Reduce result set, disable lyrics search
- **Offline**: Grace period if license server unreachable

---

## 9. Monitoring & Observability

### Key Metrics
- **Playback**: Start latency, buffering ratio, skip rate
- **Search**: Latency (p50, p99), result relevance
- **Recommendations**: Click-through rate, listen-through rate
- **API**: Latency, error rate, QPS
- **CDN**: Cache hit ratio, egress, origin load

### Alerts
- **Playback start > 2s** (p99)
- **Search latency > 500ms**
- **Error rate > 0.5%**
- **CDN cache hit < 85%**

### Tracing
- **Distributed tracing**: Request ID across services
- **Playback trace**: Auth → Playback → CDN

### Logging
- **Listen events**: Track, user, timestamp, context (for recommendations)
- **Structured logs**: JSON, searchable

---

## 10. Interview Tips

### Follow-up Questions
- "How would you design the 'Spotify Wrapped' feature (yearly summary)?"
- "How do you handle catalog differences across countries?"
- "How would you add real-time lyrics sync?"
- "How do you prevent abuse (e.g., bot streams to inflate plays)?"
- "How would you design Spotify Connect (handoff between devices)?"

### Common Mistakes
- **Treating audio like video**: Audio is smaller; chunking strategy differs
- **Ignoring licensing**: Catalog varies by region; critical for design
- **Over-complex recommendations**: Start with collaborative filtering
- **Ignoring offline**: Encrypted storage, license refresh are key
- **Single quality**: Multiple bitrates for network adaptation

### Key Points to Emphasize
- **Chunked streaming**: Pre-encoded chunks, client buffers ahead
- **Search**: Elasticsearch over metadata + lyrics; ranking with popularity
- **Recommendations**: Collaborative + content-based (audio features) + NLP
- **Offline**: Encrypted local storage, time-bound licenses
- **Licensing**: Per-territory catalog, royalty tracking

---

## Appendix: Deep Dive Topics

### A. Audio Codec Comparison
| Codec | Bitrate | Quality | Compatibility |
|-------|---------|--------|---------------|
| Ogg Vorbis 96 | 96 kbps | Low | Spotify native |
| Ogg Vorbis 128 | 128 kbps | Normal | Default |
| Ogg Vorbis 256 | 256 kbps | High | Premium |
| Ogg Vorbis 320 | 320 kbps | Very High | Premium |
| AAC | 128-256 | Good | Fallback |

### B. Search Index Structure (Elasticsearch)
- **Tracks**: title, artist_name, album_name, lyrics (analyzed)
- **Artists**: name, genres (keyword + text)
- **Albums**: title, artist_name
- **Analyzers**: Standard, lowercase, edge n-gram for autocomplete
- **Boosting**: popularity field boosts trending results

### C. Recommendation Model Types
- **CF (Collaborative Filtering)**: User-item matrix, matrix factorization
- **Content-based**: Audio features (tempo, energy, valence)
- **Sequence**: LSTM on play history for "next track"
- **Contextual bandits**: Real-time A/B for exploration

### D. Offline License Flow
1. User taps "Download" on playlist
2. Client requests license from server (auth required)
3. Server returns encrypted license (valid 30 days)
4. Client stores license + encrypted audio chunks
5. Playback: Decrypt with license key
6. Expiry: Prompt to go online; refresh license

### E. Playlist Collaboration Conflict Resolution
- **Last-write-wins**: Simple but can lose edits
- **Operational transform (OT)**: Merge concurrent edits
- **CRDT**: Conflict-free replicated data type for add/remove/reorder
- **Lock**: Single editor at a time (simplest)

### F. Audio Chunk Structure
- **Segment duration**: 20-30 seconds typical
- **Pre-buffer**: Client requests 2-3 chunks ahead
- **Range requests**: HTTP Range header for partial fetch
- **Fallback**: If chunk fails, retry; skip to next if persistent failure

### G. Catalog Territory Handling
- **Per-country**: Track availability stored as (track_id, country) → available
- **Query**: Filter by user's country when serving catalog
- **Licensing**: Expiry dates; remove when license lapses
- **Gray areas**: VPN detection; serve based on billing country

### H. Recommendation Cold Start
- **New user**: No history; use signup preferences, popular in region
- **New track**: No plays; use audio features, similar to popular tracks
- **New artist**: Rely on metadata, genre, similar artists

### I. Spotify Wrapped Design (Yearly Summary)
- **Data**: Aggregate listen history (top tracks, artists, minutes)
- **Batch job**: Run in December; compute per user
- **Storage**: Pre-computed JSON per user; cached
- **Delivery**: In-app experience; shareable cards (image generation)

### J. Spotify Connect (Device Handoff)
- **Discovery**: mDNS/Bonjour for local devices
- **Control**: Phone sends commands to speaker; speaker streams directly from CDN
- **Sync**: Playback state shared; seek on one updates all

### K. Real-Time Lyrics
- **Source**: Licensed from Musixmatch, etc.
- **Sync**: Timestamps per line; client highlights current
- **Storage**: (track_id, line, start_ms, end_ms)
- **API**: GET /tracks/:id/lyrics returns synced lines with timestamps
