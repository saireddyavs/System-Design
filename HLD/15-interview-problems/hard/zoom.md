# System Design: Zoom / Video Conferencing

## 1. Problem Statement & Requirements

### Problem Statement
Design a real-time video conferencing platform that supports 1-on-1 calls, group calls (up to 1000 participants), screen sharing, in-call chat, and meeting recording. The system must deliver low-latency, high-quality audio and video to participants globally.

### Functional Requirements
- **1-on-1 Video Calls**: Low-latency peer-to-peer or server-mediated video/audio
- **Group Calls**: Support up to 1000 participants with scalable architecture
- **Screen Sharing**: Share screen/window with participants
- **Chat**: Text chat during meetings (1-on-1 and group)
- **Recording**: Record meetings (video, audio, screen share) for later playback
- **Virtual Backgrounds**: Optional client-side processing
- **Breakout Rooms**: Split participants into sub-groups (can be phase 2)

### Non-Functional Requirements
- **Scale**: 300M daily meeting participants
- **Latency**: < 150ms end-to-end for real-time feel; < 400ms acceptable
- **Quality**: Adaptive bitrate based on network; 720p-1080p typical
- **Availability**: 99.9% for signaling; 99.95% for media
- **Security**: E2E encryption option; SRTP for media; TLS for signaling

### Security Requirements
- **Authentication**: JWT; meeting passcode optional
- **Encryption**: DTLS-SRTP for media; TLS for signaling
- **Waiting room**: Host can admit participants one-by-one
- **E2E option**: Keys only on clients; server cannot decrypt (advanced)

### Out of Scope
- Live streaming to thousands (different architecture — HLS/DASH)
- Whiteboard/collaborative canvas
- Calendar integration (OAuth, sync)
- Mobile-specific optimizations (battery, thermal)

---

## 2. Back-of-Envelope Estimation

### Capacity Planning Summary

Video conferencing is bandwidth and compute intensive. Key drivers: concurrent meetings, participants per meeting, video quality, recording percentage.

### Assumptions
- 300M daily participants
- Average meeting: 4 participants, 45 min
- Peak concurrent meetings: 10M (assuming 8-hour workday spread)
- Video: 720p @ 1.5 Mbps; Audio: 64 kbps per participant

### QPS Estimates
| Operation | Daily Volume | QPS (peak) |
|-----------|--------------|------------|
| Meeting joins | 300M | ~3,500 |
| Signaling (SDP, ICE) | 1.2B | ~14,000 |
| Media packets | Trillions | N/A (handled by media servers) |

### Bandwidth Estimates (Peak)
- **Per participant upload**: 1.5 Mbps video + 64 kbps audio ≈ 1.6 Mbps
- **Per participant download** (N-1 others): (N-1) × 1.6 Mbps
- **1-on-1**: 2 × 1.6 = 3.2 Mbps per call
- **10-person call**: 10 × 9 × 1.6 = 144 Mbps per call
- **With SFU**: Each sends 1 stream; SFU receives 10, sends 90 (9 per participant)
- **Total ingress to SFU fleet**: 10M meetings × avg 4 × 1.6 Mbps ≈ 64 Tbps (simplified)
- **Realistic**: Many 1-on-1 (P2P), small groups; estimate 20-30 Tbps peak

### Storage (Recording)
- 10% of meetings recorded; avg 45 min
- 720p @ 1.5 Mbps + audio ≈ 2 Mbps = 90 MB per 45 min per recording
- 30M recordings × 90 MB = 2.7 PB/day (raw); with compression ~1 PB/day

### Cache
- **Signaling**: Session state in Redis; TTL = meeting duration
- **Recording metadata**: DB; actual video in object storage

### Cost Considerations

- **TURN**: Most expensive (relays all media); minimize % of traffic using TURN
- **SFU**: Moderate (receives and forwards; no transcode); scales with participants
- **Storage**: Recording dominates; lifecycle policies (delete after 90 days) reduce cost

---

## 3. API Design

### REST Endpoints

```
# Authentication
POST   /api/v1/auth/login          # OAuth or email/password
POST   /api/v1/auth/refresh        # Refresh token
GET    /api/v1/users/me            # Current user

# Meetings
POST   /api/v1/meetings            # Create meeting (returns meeting_id, passcode)
GET    /api/v1/meetings/{id}       # Get meeting info
POST   /api/v1/meetings/{id}/join  # Join meeting (returns signaling URL, TURN creds)
POST   /api/v1/meetings/{id}/leave # Leave meeting
DELETE /api/v1/meetings/{id}       # End meeting (host only)

# Recording
POST   /api/v1/meetings/{id}/recording/start   # Start recording
POST   /api/v1/meetings/{id}/recording/stop    # Stop recording
GET    /api/v1/recordings/{id}     # Get recording metadata
GET    /api/v1/recordings/{id}/download  # Signed URL for download

# Chat
POST   /api/v1/meetings/{id}/chat  # Send chat message
GET    /api/v1/meetings/{id}/chat  # Get chat history (paginated)
```

### WebSocket (Signaling)

```
WS /api/v1/signaling/{meeting_id}
- Connect with auth token
- Messages (JSON):
  - offer: { type: "offer", sdp: "...", from: "user_id" }
  - answer: { type: "answer", sdp: "...", from: "user_id" }
  - ice-candidate: { type: "ice", candidate: "...", from: "user_id" }
  - join: { type: "join", user_id, display_name }
  - leave: { type: "leave", user_id }
  - screen-share-start: { type: "screen_start", from: "user_id" }
  - screen-share-stop: { type: "screen_stop", from: "user_id" }
```

### TURN/STUN Credentials
```
GET /api/v1/turn/credentials
Response: {
  "urls": ["stun:stun.zoom.com", "turn:turn.zoom.com"],
  "username": "temp_user_123",
  "credential": "temp_password_456",
  "ttl": 86400
}
```

---

## 4. Data Model / Database Schema

### Database Choice
- **Metadata**: PostgreSQL — meetings, users, recordings
- **Session state**: Redis — active meeting participants, signaling state
- **Recording storage**: S3/GCS — video files

### Schema

**users**
| Column | Type | Description |
|--------|------|-------------|
| user_id | UUID PK | |
| email | VARCHAR | |
| display_name | VARCHAR | |
| avatar_url | VARCHAR | |
| created_at | TIMESTAMP | |

**meetings**
| Column | Type | Description |
|--------|------|-------------|
| meeting_id | UUID PK | |
| host_id | UUID FK | |
| title | VARCHAR | |
| scheduled_at | TIMESTAMP | |
| duration_minutes | INT | |
| passcode | VARCHAR | Optional |
| max_participants | INT | Default 1000 |
| created_at | TIMESTAMP | |

**meeting_participants** (ephemeral, can use Redis)
| Column | Type | Description |
|--------|------|-------------|
| meeting_id | UUID | |
| user_id | UUID | |
| joined_at | TIMESTAMP | |
| left_at | TIMESTAMP | Null if still in call |
| role | ENUM | host, participant |

**recordings**
| Column | Type | Description |
|--------|------|-------------|
| recording_id | UUID PK | |
| meeting_id | UUID FK | |
| storage_path | VARCHAR | S3 key |
| duration_seconds | INT | |
| status | ENUM | processing, ready, failed |
| created_at | TIMESTAMP | |

**chat_messages**
| Column | Type | Description |
|--------|------|-------------|
| message_id | UUID PK | |
| meeting_id | UUID FK | |
| user_id | UUID FK | |
| content | TEXT | |
| created_at | TIMESTAMP | |

### Indexes

- `meetings(host_id)` — User's meetings
- `meeting_participants(meeting_id)` — Who is in call
- `recordings(meeting_id)` — Meeting recordings
- `chat_messages(meeting_id, created_at)` — Chat history

### Redis Keys (Session State)

- `meeting:{id}:participants` — Set of user_ids in meeting
- `meeting:{id}:host` — Host user_id
- `user:{id}:meeting` — Current meeting_id for user (for reconnect)

---

## 5. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              CLIENTS (Web, Desktop, Mobile)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │ WebRTC       │  │ Media Capture│  │ Audio Proc   │  │ Signaling Client     │ │
│  │ (getUserMedia│  │ (camera,     │  │ (AGC, AEC,   │  │ (WebSocket)           │ │
│  │  RTCPeerConn)│  │  mic, screen)│  │  noise cancel│  │                       │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────────┬─────────────┘ │
└─────────┼─────────────────┼─────────────────┼────────────────────┼──────────────┘
          │                 │                 │                    │
          │    SRTP/RTP      │                 │                    │ WebSocket
          │    (media)       │                 │                    │ (signaling)
          ▼                 ▼                 ▼                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              EDGE / LOAD BALANCER                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
          │                                    │
          ▼                                    ▼
┌─────────────────────────────┐    ┌─────────────────────────────────────────────┐
│  SIGNALING SERVER            │    │  MEDIA SERVERS (SFU / MCU)                   │
│  - WebSocket connections     │    │  - Receive streams from participants        │
│  - SDP exchange              │    │  - Forward to other participants (SFU)      │
│  - ICE candidate relay       │    │  - Or mix and send (MCU)                     │
│  - Meeting state             │    │  - Recording ingestion                      │
└─────────────────────────────┘    └─────────────────────────────────────────────┘
          │                                    │
          ▼                                    ▼
┌─────────────────┐                 ┌─────────────────┐  ┌─────────────────┐
│ Redis           │                 │ Object Storage  │  │ Post-processing │
│ (session state) │                 │ (recordings)    │  │ (transcode,     │
└─────────────────┘                 └─────────────────┘  │  thumbnail)      │
                                                         └─────────────────┘

P2P (1-on-1)                    SFU (Group)                      MCU (Alternative)
┌──────┐      ┌──────┐          ┌──────┐  ┌──────┐  ┌──────┐     ┌──────┐  ┌──────┐
│Client│◄────►│Client│          │Client│─►│ SFU  │─►│Client│     │Client│─►│ MCU  │
│  A   │      │  B   │          │  A   │  │      │  │  B   │     │  A   │  │      │
└──────┘      └──────┘          └──────┘  │  ◄───┼──│  C   │     │  B   │  │ Mix  │
     Direct connection           Each     │      │  └──────┘     │  C   │  │ 1    │
     via STUN/TURN               sends 1  └──────┘               └──┬───┘  │stream│
     (no server in path)         stream   O(N) ingress            │       └──────┘
     if NAT allows               SFU       O(N²) egress           │       Each gets
     forwards                    (N streams)                     │       single mixed
                                                                  └──────► stream
```

---

## 6. Detailed Component Design

### 6.1 WebRTC Fundamentals

**WebRTC** enables peer-to-peer media (audio, video, data) in browsers without plugins.

**STUN (Session Traversal Utilities for NAT)**:
- Client sends request to STUN server; server returns client's public IP:port
- Used to discover public address when behind NAT
- Does not relay media

**TURN (Traversal Using Relays around NAT)**:
- When P2P fails (symmetric NAT, firewall), TURN relays media
- Client sends media to TURN; TURN forwards to peer
- Expensive (bandwidth); use as fallback

**ICE (Interactive Connectivity Establishment)**:
- Gathers candidate pairs (host, server reflexive, relay)
- Tests connectivity; selects best pair
- Ensures connection works across NAT/firewall

**SDP (Session Description Protocol)**:
- Describes media (codecs, resolution, etc.)
- Offer/Answer model: one side sends offer, other answers
- Exchanged via signaling channel (WebSocket)

### 6.2 P2P vs SFU vs MCU

**P2P (1-on-1)**:
- Direct connection between two clients
- Lowest latency, no server bandwidth
- Fails if NAT traversal fails → use TURN
- Does not scale to 3+ participants (each would need N-1 connections)

**SFU (Selective Forwarding Unit)**:
- Each participant sends ONE stream to SFU
- SFU forwards to other participants (does not decode/mix)
- Ingress: N streams (one per participant)
- Egress: N × (N-1) streams (each participant gets N-1 streams)
- O(N) server ingress, O(N²) egress — but egress is to clients, distributed
- **Simulcast**: Sender produces 3 quality layers (low, mid, high); SFU picks per recipient based on bandwidth
- **Scalable**: SFU does minimal work; no transcoding

**MCU (Multipoint Control Unit)**:
- Server receives all streams, decodes, mixes into one
- Sends single mixed stream to each participant
- O(N) egress (one stream per participant)
- High server CPU (decode N, encode 1 per participant)
- Lower client bandwidth (one stream)
- Used for legacy/hardware; less common in software

### 6.3 Signaling Server

**Responsibilities**:
- Maintain WebSocket connection per participant
- Relay SDP offer/answer between participants
- Relay ICE candidates
- Broadcast join/leave events
- Manage meeting state (who is in, who is sharing screen)

**Flow**:
1. Client A creates meeting, gets meeting_id
2. Client B joins, connects WebSocket to signaling server
3. For P2P: A sends offer to B via server; B sends answer to A
4. A and B exchange ICE candidates via server
5. Once connected, media flows P2P (or via TURN)

**For SFU**:
1. Each client sends offer to SFU (or SFU sends offer to client)
2. SFU maintains N peer connections (one per participant)
3. SFU forwards each incoming stream to other participants

### 6.4 Media Server (SFU) Selection

**Geographic distribution**: Deploy SFU nodes in multiple regions (US, EU, Asia).

**Selection**: Client reports location or measures latency to candidate SFU nodes. Assign meeting to SFU closest to majority of participants. Use DNS or API to return SFU URL.

**Scaling**: Each SFU handles 100-500 concurrent participants. Auto-scale based on CPU/bandwidth.

### 6.5 Adaptive Bitrate & Simulcast

**Adaptive bitrate**: Client or SFU adjusts quality based on network conditions.
- **Simulcast**: Sender encodes 3 layers (e.g., 180p, 360p, 720p). SFU subscribes to appropriate layer per receiver.
- **SFU-driven**: SFU sends RTCP feedback (e.g., "reduce bitrate"); sender adapts.
- **Receiver-driven**: Receiver requests different simulcast layer via signaling.

### 6.6 Audio Processing

**Client-side** (before sending):
- **AGC (Automatic Gain Control)**: Normalize volume
- **AEC (Acoustic Echo Cancellation)**: Remove echo from speakers
- **Noise suppression**: Reduce background noise (e.g., keyboard, fan)
- **VAD (Voice Activity Detection)**: Detect speech; can reduce bandwidth when silent

**Libraries**: WebRTC includes built-in AGC, AEC. For advanced: RNNoise, Speex.

### 6.7 Recording Service

**Architecture**:
1. **SFU/MCU** sends copy of media to recording service (or recording runs on same media server)
2. **Recorder** receives RTP streams, writes to container (e.g., WebM, MP4)
3. **Upload** to object storage (S3) when meeting ends
4. **Post-processing**: Transcode to standard format (H.264), generate thumbnail, extract chat
5. **Metadata** stored in DB; signed URL for playback

**Challenges**: Sync audio/video from multiple participants; layout (grid vs active speaker).

### 6.8 SRTP (Secure Real-time Transport Protocol)

- Encrypts RTP media (audio, video)
- Uses keys derived from DTLS handshake (part of WebRTC)
- E2E encryption: Keys only on clients; server cannot decrypt (SFU forwards encrypted packets without decoding)

### 6.9 Large Meeting Mode (1000+ Participants)

**Challenge**: Sending 999 video streams to each participant is infeasible (bandwidth, CPU).

**Strategies**:
- **Active speaker only**: SFU forwards only the current speaker's video (and maybe 1-2 previous). Others get audio only or thumbnail.
- **Spotlight**: Host pins 1-9 participants; others see only those.
- **Gallery limit**: Client requests top N (e.g., 25) by "importance" (speaking, pinned, host).
- **Simulcast**: Sender produces low-res thumbnail; SFU sends thumbnail for non-active participants.
- **Server-side layout**: MCU-style mixing for "broadcast" view (higher server cost).

### 6.10 Chat Architecture

**In-meeting chat**:
- **Signaling channel**: Send chat over WebSocket (same as signaling) or separate channel
- **Persistence**: Store in DB for replay; optional for small meetings
- **Scale**: 1000 participants × 10 msg/min = 10K msg/min; Redis pub/sub or Kafka for fan-out
- **Ordering**: Lamport timestamp or sequence number per meeting

**Private chat**: 1-on-1; same as above but filter by recipient.

### 6.11 Network Adaptation (Detailed)

**RTCP feedback**: Receiver sends Receiver Reports (RR) with packet loss, jitter. Sender adjusts.

**PLI (Picture Loss Indication)**: Receiver requests keyframe when it detects gap. Sender sends I-frame.

**Fir (Full Intra Request)**: Similar to PLI; request full frame.

**Bitrate adaptation**: 
- **Aimd**: Additive increase, multiplicative decrease (like TCP)
- **Google Congestion Control (GCC)**: Uses delay-based and loss-based signals
- **Simulcast switch**: SFU or receiver requests lower layer when congestion detected

### 6.12 Screen Sharing Optimization

**Challenge**: Screen content changes less frequently than camera; different encoding strategy.

**Approaches**:
- **Lower frame rate**: 5-10 fps often sufficient
- **Region of interest**: Encode only changed region (partial frame)
- **Higher compression**: Screen content compresses well (text, solid colors)
- **Separate stream**: Different SDP track; SFU forwards to participants who opted in

---

## 7. Scaling

### Signaling

- **Horizontal scaling**: Stateless signaling servers; session state in Redis
- **WebSocket**: Each connection = 1 server; use sticky sessions (load balancer)
- **Redis pub/sub**: Broadcast join/leave to all participants in meeting

### Media (SFU)

- **Per-meeting SFU**: One SFU instance per meeting (or per N meetings)
- **Regional deployment**: SFUs in each region; route by latency
- **Auto-scaling**: Scale SFU fleet by concurrent participants

### TURN

- **TURN servers**: Deploy globally; high bandwidth
- **Credential rotation**: Short-lived TURN credentials (e.g., 24h) to prevent abuse

### Recording

- **Async processing**: Record to local disk; async upload to S3
- **Transcoding**: Worker pool for post-processing; queue-based

### Geographic Routing

- **DNS-based**: `us-meetings.zoom.com` vs `eu-meetings.zoom.com` — user routed by region
- **Anycast**: Single hostname; BGP routes to nearest PoP
- **Client hint**: Client sends region/country; API returns nearest signaling + SFU URLs

### Connection Limits

- **Per-SFU**: 400-500 participants (empirically); beyond that, split meeting across SFUs (complex)
- **Per-signaling**: 10K WebSocket connections per server (memory-bound)
- **TURN**: 5-10 Gbps per server; scale horizontally

---

## 8. Failure Handling

### Component Failures

| Component | Failure | Mitigation |
|-----------|---------|------------|
| Signaling | Server crash | Reconnect WebSocket; restore state from Redis |
| SFU | Node failure | Migrate participants to backup SFU; brief interruption |
| TURN | Overload | Multiple TURN servers; fallback to different region |
| Recording | Disk full | Alert; fail recording gracefully; retry upload |

### Redundancy

- **Signaling**: Multiple instances; Redis for state
- **SFU**: Active-passive or active-active per region
- **TURN**: Multiple servers per region

### Graceful Degradation

- **Video fails**: Continue audio-only
- **Network congestion**: Reduce resolution/bitrate
- **SFU overload**: Reject new joins; notify host

### Reconnection Flow

1. Client detects disconnect (WebSocket closed, media stopped)
2. Reconnect WebSocket to signaling server
3. Request meeting state (participants, who is sharing)
4. Re-establish WebRTC to SFU (new offer/answer)
5. Resume media flow. Brief gap (1-5s) acceptable.

### Recording Failure Recovery

- Record to local disk first; upload async. If upload fails, retry with exponential backoff.
- Store partial recording if meeting ends abruptly; mark as "incomplete" in metadata.

---

## 9. Monitoring & Observability

### Key Metrics

| Metric | Description | Alert |
|--------|-------------|-------|
| signaling_connection_count | Active WebSockets | Capacity |
| sfu_participant_count | Participants per SFU | > 400 |
| sfu_cpu_usage | CPU per SFU | > 80% |
| turn_bandwidth | TURN relay bandwidth | Capacity |
| join_latency | Time to join meeting | > 5s |
| media_latency_p99 | RTT for media | > 300ms |
| recording_failure_rate | Failed recordings | > 1% |

### Logging

- Meeting lifecycle: create, join, leave, end
- Signaling: offer/answer/ICE exchange (log IDs, not SDP)
- Errors: connection failures, SFU errors

### Tracing

- Trace join flow: API → signaling → WebRTC connection
- Trace recording: start → ingest → upload → transcode

### SLOs

- **Join success rate**: 99.5% (excluding user network issues)
- **Media latency**: p99 < 300ms RTT
- **Recording success**: 99% of requested recordings complete

### Dashboards

- **Real-time**: Active meetings, participants, SFU load, TURN bandwidth
- **Quality**: Join latency, media latency, packet loss by region
- **Business**: DAU, meeting minutes, recording storage

---

## 10. Interview Tips

### Follow-up Questions

1. **How do you handle 1000 participants?** SFU with simulcast; each receives subset (e.g., active speakers); "spotlight" mode for large meetings.
2. **What if TURN is overloaded?** Multiple TURN servers; geographic distribution; quota per user.
3. **How does screen sharing work?** Same as video; separate MediaStream; can use lower bitrate or region-of-interest encoding.
4. **E2E encryption with SFU?** SFU forwards encrypted packets without decrypting; keys exchanged between clients (more complex signaling).
5. **How do you reduce bandwidth for "audio-only" participants?** Simulcast; send only audio layer; or SFU stops forwarding video to them.

### Common Mistakes

1. **P2P for group calls**: Does not scale; need SFU/MCU.
2. **Ignoring TURN**: Many corporate networks block P2P; TURN is essential.
3. **Single region**: Latency for global users; need edge deployment.
4. **No simulcast**: Cannot adapt to varying client bandwidth.
5. **Recording on client**: Unreliable; record on server.

### What to Emphasize

- **SFU vs MCU trade-offs**: SFU scales better; MCU reduces client bandwidth
- **Simulcast**: Key for adaptive quality
- **STUN/TURN/ICE**: Essential for NAT traversal
- **Signaling vs media**: Separate channels; media is bulk of traffic
- **Geographic distribution**: Critical for latency

### Sample Discussion Flow

1. **Clarify**: "1-on-1 or group? Recording?" — Both; cover P2P for 1-on-1, SFU for group.
2. **WebRTC basics**: STUN discovers public IP; TURN relays when P2P fails; ICE picks best path.
3. **Why SFU for group?** P2P needs N×(N-1)/2 connections; SFU needs N ingress, N×(N-1) egress (distributed).
4. **Simulcast**: "Variable bandwidth?" → Sender sends 3 layers; SFU picks per receiver.
5. **Recording**: Server records; upload to S3; transcode async.

### Time-Boxed Approach (45 min interview)

- **0-5 min**: Clarify 1-on-1 vs group, recording, scale
- **5-15 min**: High-level (signaling, media path, P2P vs SFU)
- **15-25 min**: WebRTC (STUN/TURN/ICE), SFU architecture, simulcast
- **25-35 min**: Scaling, recording, failure handling
- **35-45 min**: Large meetings, E2E encryption, follow-ups

### Additional Deep-Dive Topics

**Codec selection**:
- **Video**: VP8/VP9 (WebRTC default), H.264 (hardware support). H.264 often better for compatibility.
- **Audio**: Opus (low latency, good quality). Fallback: G.711 for legacy.

**TURN allocation**:
- TURN server allocates relay address per client. Each allocation consumes ports. Limit per user to prevent abuse.

**Meeting migration**:
- SFU fails mid-call: Signaling server detects; instructs clients to reconnect to new SFU. Brief audio/video gap.

**Bandwidth calculation (per participant)**:
- 720p: ~1.5 Mbps; 1080p: ~3 Mbps; Audio: 64 kbps. Screen share: 1-2 Mbps.

### Design Alternatives Considered

**P2P vs SFU for 3-person call**: P2P needs 3 connections per client (mesh). SFU needs 1 upload, 2 downloads. At 3, similar. SFU scales better; use SFU for 3+.

**MCU vs SFU**: MCU reduces client bandwidth (1 stream) but needs server CPU to decode/mix. SFU scales horizontally; MCU does not. SFU for software-first design.

**Recording: client vs server**: Client recording unreliable (user closes tab, network). Server recording guaranteed. Server wins.

**Signaling: REST vs WebSocket**: REST for one-off (create meeting). WebSocket for ongoing (offer/answer/ICE). Hybrid.

### Phased Rollout

**Phase 1 (MVP)**: 1-on-1 P2P only. STUN/TURN. Basic signaling. No recording.

**Phase 2**: Group calls with SFU. Simulcast. Chat. Up to 50 participants.

**Phase 3**: Recording. Large meeting mode (active speaker). 100+ participants.

**Phase 4**: 1000 participants. E2E encryption. Breakout rooms. Global scale.

### Quick Reference Card

| Concept | Key Point |
|---------|-----------|
| 1-on-1 | P2P via STUN/TURN/ICE |
| Group | SFU; each sends 1, receives N-1 |
| STUN | Discover public IP |
| TURN | Relay when P2P fails |
| Simulcast | 3 layers; SFU picks per receiver |
| Recording | Server-side; SFU → recorder → S3 |
| Signaling | WebSocket; SDP + ICE exchange |
| Latency | < 150ms target; < 400ms acceptable |

### Glossary

- **SFU**: Selective Forwarding Unit — forwards streams without mixing
- **MCU**: Multipoint Control Unit — mixes streams server-side
- **SDP**: Session Description Protocol — describes media (codecs, etc.)
- **ICE**: Interactive Connectivity Establishment — NAT traversal
- **Simulcast**: Sender produces multiple quality layers

### Codec Details

- **Video**: H.264 (Baseline/Main) for compatibility; VP8/VP9 for WebRTC default. H.264 has hardware encode/decode on most devices.
- **Audio**: Opus (default); 48kHz, 32-64 kbps. Low latency, good quality. Fallback: G.711 for PSTN.
- **Container**: WebM for recording (VP8+Opus); MP4 for playback (H.264+AAC) after transcode.

### NAT Types and TURN Usage

- **Full cone / Restricted cone / Port restricted**: STUN usually sufficient; P2P works.
- **Symmetric NAT**: P2P often fails; TURN required. ~10-20% of users behind symmetric NAT.
- **Corporate firewall**: Often blocks UDP; TURN over TCP or TURN over TLS needed.

### Meeting Lifecycle

1. **Create**: Host calls API; gets meeting_id, passcode. Meeting record in DB.
2. **Join**: Participant connects WebSocket; sends join. Signaling relays to others. WebRTC established.
3. **In-call**: Media flows; chat optional. Host can mute, remove participants.
4. **Leave**: Participant disconnects; signaling broadcasts leave. Resources released.
5. **End**: Host ends; all disconnected. Meeting marked ended. Recording finalized if was on.

### Bandwidth per Participant (Approximate)

| Scenario | Upload | Download |
|----------|--------|----------|
| 1-on-1 (720p) | 1.5 Mbps | 1.5 Mbps |
| 10-person (720p) | 1.5 Mbps | 13.5 Mbps |
| 100-person (active speaker) | 1.5 Mbps | 1.5 Mbps |
| Screen share | 1-2 Mbps | 1-2 Mbps per viewer |

*Note: Download scales with number of video streams received. Active-speaker mode limits to 1-2 streams per participant.*

**Audio-only mode**: ~64 kbps per participant; significantly reduces bandwidth for large meetings. Useful when video is not critical.

---
