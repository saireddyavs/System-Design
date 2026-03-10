# Design Google Maps

## 1. Problem Statement & Requirements

### Problem Statement
Design a mapping and navigation platform that renders maps, provides directions (driving, walking, transit), shows real-time traffic, calculates ETAs, supports place search, and offers geolocation services at global scale.

### Functional Requirements
- **Map rendering**: Display maps at various zoom levels; support vector and raster tiles
- **Directions**: Driving, walking, cycling, transit (multi-modal)
- **Real-time traffic**: Current traffic conditions; color-coded (green/yellow/red)
- **ETA**: Accurate arrival time estimates
- **Place search**: Search by name, address, category; autocomplete
- **Geocoding**: Address → coordinates
- **Reverse geocoding**: Coordinates → address
- **Geolocation**: User's current position
- **Points of interest (POI)**: Restaurants, gas stations, etc.

### Non-Functional Requirements
- **Latency**: Map tile load < 200 ms; routing < 2 seconds
- **Availability**: 99.99%
- **Scale**: 1B+ monthly users; petabytes of map data
- **Global coverage**: Worldwide maps and routing

### Out of Scope
- Street View (simplified)
- Real-time incident reporting (Waze-style)
- Indoor maps
- AR navigation

---

## 2. Back-of-Envelope Estimation

### Assumptions
- 1B monthly users; 100M daily active
- 50M direction requests/day
- 500M map tile requests/day
- 200M place searches/day
- 50M geocoding requests/day

### QPS Estimates
| Component | Calculation | QPS |
|-----------|-------------|-----|
| Map tiles | 500M / 86400 | ~5,800 |
| Peak tiles (5x) | 5800 × 5 | ~29,000 |
| Directions | 50M / 86400 | ~580 |
| Place search | 200M / 86400 | ~2,300 |
| Geocoding | 50M / 86400 | ~580 |

### Storage
| Data | Size | Notes |
|------|------|-------|
| Road network graph | 100+ TB | Nodes, edges, attributes |
| Tile pyramid (raster) | 1–10 PB | All zoom levels, global |
| Vector tiles | 500 TB | Compact, styleable |
| POI database | 50 TB | Places, reviews, hours |
| Traffic data | 10 TB/day | Segment speeds, historical |

### Bandwidth
- Tile delivery: 29K QPS × 50 KB ≈ 1.5 GB/s
- CDN handles majority

### Cache
- Tile cache: CDN + edge; 90%+ hit rate
- Route cache: (origin, dest) → route; 50% hit rate
- Geocode cache: Address → coords; 80% hit rate

---

## 3. API Design

### REST Endpoints

```
GET  /api/v1/maps/tiles/{z}/{x}/{y}           # Get map tile (raster/vector)
GET  /api/v1/directions                      # Get directions
GET  /api/v1/geocode                         # Geocode address
GET  /api/v1/geocode/reverse                  # Reverse geocode
GET  /api/v1/places/search                   # Place search
GET  /api/v1/places/autocomplete             # Autocomplete
GET  /api/v1/places/{placeId}                # Place details
GET  /api/v1/traffic/segment/{segmentId}     # Traffic for segment
GET  /api/v1/eta                             # ETA from A to B
```

### Request/Response Examples

**Directions**
```
GET /api/v1/directions?origin=37.77,-122.42&destination=37.78,-122.41&mode=driving

Response:
{
  "routes": [{
    "legs": [{
      "distance": {"value": 5000, "text": "5 km"},
      "duration": {"value": 600, "text": "10 mins"},
      "duration_in_traffic": {"value": 900, "text": "15 mins"},
      "steps": [...]
    }],
    "overview_polyline": "encoded_string"
  }]
}
```

**Place Search**
```
GET /api/v1/places/search?query=coffee&location=37.77,-122.42&radius=5000

Response:
{
  "results": [{
    "place_id": "ChIJ...",
    "name": "Blue Bottle Coffee",
    "geometry": {"location": {"lat": 37.77, "lng": -122.42}},
    "vicinity": "66 Mint St, San Francisco"
  }]
}
```

---

## 4. Data Model / Database Schema

### Map Tile System

**Tile pyramid** (Web Mercator, zoom levels 0–22):
- Zoom 0: 1 tile (whole world)
- Zoom 1: 4 tiles
- Zoom n: 4^n tiles
- Zoom 22: billions of tiles

**Tile addressing**: (z, x, y) — zoom, column, row

### Road Network Graph

**nodes**
| Column | Type |
|--------|------|
| id | BIGINT PK |
| lat | DOUBLE |
| lng | DOUBLE |
| elevation | FLOAT |

**edges**
| Column | Type |
|--------|------|
| id | BIGINT PK |
| from_node | BIGINT FK |
| to_node | BIGINT FK |
| length_m | FLOAT |
| road_type | ENUM |
| speed_limit_kmh | INT |
| one_way | BOOLEAN |
| name | VARCHAR |

**turn_restrictions**
| Column | Type |
|--------|------|
| from_edge | BIGINT |
| via_node | BIGINT |
| to_edge | BIGINT |

### Places/POI

**places**
| Column | Type |
|--------|------|
| id | VARCHAR PK |
| name | VARCHAR |
| lat | DOUBLE |
| lng | DOUBLE |
| type | VARCHAR |
| address | TEXT |
| metadata | JSONB |

### DB Choice
- **PostgreSQL + PostGIS**: Geospatial queries, routing
- **Elasticsearch**: Place search, autocomplete
- **Redis**: Tile cache, route cache, geocode cache
- **HBase/Cassandra**: Traffic time-series

---

## 5. High-Level Architecture

```
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                    CDN (CloudFront/Cloudflare)              │
                                    │              (Tiles, Static Assets, Cached Routes)          │
                                    └─────────────────────────────────────────────────────────────┘
                                                              │
                                                              │ Cache miss
                                                              ▼
                                    ┌─────────────────────────────────────────────────────────────┐
                                    │                      LOAD BALANCER                           │
                                    └─────────────────────────────────────────────────────────────┘
                                                              │
         ┌────────────────────────────────────────────────────┼────────────────────────────────────────────────────┐
         │                                                    │                                                    │
         ▼                                                    ▼                                                    ▼
┌─────────────────┐                              ┌─────────────────┐                              ┌─────────────────┐
│  Tile Server    │                              │  Routing Service│                              │  Search Service │
│  (Vector/Raster)│                              │  (Directions)   │                              │  (Places)       │
└────────┬────────┘                              └────────┬────────┘                              └────────┬────────┘
         │                                                │                                                │
         │                                                │                                                │
         ▼                                                ▼                                                ▼
┌─────────────────┐                              ┌─────────────────┐                              ┌─────────────────┐
│  Tile Storage   │                              │  Routing Engine │                              │  Elasticsearch │
│  (S3/GCS)       │                              │  (CH, A*)       │                              │  (Places index) │
└─────────────────┘                              └────────┬────────┘                              └─────────────────┘
                                                          │                                                │
         ┌────────────────────────────────────────────────┼────────────────────────────────────────────────┘
         │                                                │
         ▼                                                ▼
┌─────────────────┐                              ┌─────────────────┐
│  Geocoding      │                              │  Traffic Service│
│  Service        │                              │  (Real-time)   │
└────────┬────────┘                              └────────┬────────┘
         │                                                │
         │                                                ▼
         │                                       ┌─────────────────┐
         │                                       │  Traffic Store  │
         │                                       │  (Segment speed)│
         │                                       └─────────────────┘
         │
         └───────────────────────────────────────────────────────────────┐
                                                                         │
         ┌───────────────────────────────────────────────────────────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  PostgreSQL     │     │  Redis          │     │  Kafka           │
│  + PostGIS      │     │  (Caches)       │     │  (GPS pipeline)  │
│  (Road graph)   │     │                 │     │                  │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

---

## 6. Detailed Component Design

### 6.1 Map Tile System

**Tile pyramid**:
- **Zoom 0**: 256×256 px tile of entire world
- **Zoom n**: 2^n × 2^n tiles
- **Web Mercator**: Projection for web maps

**Raster vs vector tiles**:

| Aspect | Raster | Vector |
|--------|--------|--------|
| Format | PNG/JPEG | Protocol Buffers (Mapbox Vector Tiles) |
| Size | 10–50 KB/tile | 2–10 KB/tile |
| Styling | Fixed | Client-side (change colors, labels) |
| Rendering | Server | Client (GPU) |
| Zoom | One set per zoom | Single set, client zooms |

**Tile server flow**:
1. Client requests `/tiles/{z}/{x}/{y}.png`
2. CDN checks cache; miss → origin
3. Tile server: check Redis → render or fetch from storage
4. Return tile; CDN caches

**Tile storage**:
- S3/GCS: Key = `{z}/{x}/{y}.pbf`
- Pre-generated for popular zoom levels
- On-demand generation for rare tiles

### 6.2 Routing Algorithm

**Graph representation**:
- Nodes: Intersections, road endpoints
- Edges: Road segments with length, speed limit, road type
- Turn restrictions: Forbidden turns

**Algorithms**:

1. **Dijkstra**: Shortest path; explores uniformly
2. **A***: Heuristic (e.g., straight-line distance to dest); fewer nodes explored
3. **Contraction Hierarchies (CH)**: Preprocessing adds shortcut edges; query is much faster (milliseconds for continental routes)

**Contraction Hierarchies (simplified)**:
- Order nodes by importance
- "Contract" least important nodes: add shortcuts so distances preserved
- Query: Bidirectional search from origin and destination; meet at high-level nodes

**Multi-modal transit**:
- Separate graph: transit stops, connections, schedules
- Or: Time-expanded graph (nodes = (stop, time))
- Algorithm: Dijkstra with time-dependent edge weights

**Turn restrictions**:
- Store (from_edge, via_node, to_edge) as forbidden
- During routing: when at via_node, disallow transition to to_edge

### 6.3 Real-Time Traffic

**Data sources**:
- GPS probes from phones (anonymized, aggregated)
- Third-party traffic data
- Historical patterns

**Pipeline**:
1. Ingest GPS points (lat, lng, speed, timestamp)
2. Map-matching: Snap point to road segment
3. Aggregate: Per segment, compute average speed over 5–15 min window
4. Store: segment_id → speed, timestamp
5. Serve: Traffic API returns segment speeds

**Storage**:
- Time-series DB (InfluxDB, TimescaleDB) or HBase
- Key: (segment_id, time_bucket)
- Value: avg_speed, sample_count

**Traffic prediction**:
- Historical: Same time of day, day of week
- Real-time: Blend current speed with historical
- ML: LSTM/transformer for short-term prediction

### 6.4 ETA Calculation

- Use routing engine with traffic-weighted edges
- Edge weight = length / speed (where speed = min(speed_limit, traffic_speed))
- Sum weights along path → travel time
- Add buffer for uncertainty (e.g., 10%)

### 6.5 Geocoding / Reverse Geocoding

**Geocoding** (address → coordinates):
- Parse address: street, city, country
- Search in address database (Elasticsearch, PostGIS)
- Return best match with confidence score
- Cache: address → (lat, lng)

**Reverse geocoding** (coordinates → address):
- Find nearest road segment
- Interpolate address along segment
- Or: Point-in-polygon for administrative boundaries
- Cache: (rounded lat, lng) → address

### 6.6 Place Search

**Index** (Elasticsearch):
- Fields: name, type, address, location (geo_point)
- Analyzers: N-gram for autocomplete, standard for search

**Geospatial query**:
- Filter: `geo_distance` from user location
- Or: `geo_bounding_box`
- Rank by: distance + text relevance

**Spatial index** (alternative):
- R-tree or quadtree for POI
- Query: Range or nearest-neighbor

**Autocomplete**:
- Prefix search on name
- Use completion suggester or edge n-grams
- Limit to nearby (e.g., 50 km)

### 6.7 Map Data Pipeline

**Sources**:
- OpenStreetMap (OSM)
- Commercial providers (HERE, TomTom)
- Satellite imagery
- User contributions

**Pipeline**:
1. **Ingest**: OSM XML/Protobuf
2. **Parse**: Extract nodes, ways, relations
3. **Transform**: Build road graph, extract POI
4. **Validate**: Topology, connectivity
5. **Load**: PostgreSQL, tile generation
6. **Publish**: Tiles to CDN

**Tile generation**:
- Tippecanoe (Mapbox): Vector tiles from GeoJSON
- GDAL: Raster tiles
- Batch jobs: Generate all zoom levels for region

### 6.8 ASCII Diagram: Tile Pyramid

```
Zoom 0:                    Zoom 1:                    Zoom 2:
┌─────────┐                ┌─────┬─────┐              ┌───┬───┬───┬───┐
│         │                │  0  │  1  │              │ 0 │ 1 │ 2 │ 3 │
│    1    │       →        ├─────┼─────┤      →      ├───┼───┼───┼───┤
│  tile   │                │  2  │  3  │              │ 4 │ 5 │ 6 │ 7 │
│         │                └─────┴─────┘              ├───┼───┼───┼───┤
└─────────┘                                          │ 8 │ 9 │10 │11 │
                                                     ├───┼───┼───┼───┤
                                                     │12 │13 │14 │15 │
                                                     └───┴───┴───┴───┘
```

### 6.9 ASCII Diagram: Routing Graph

```
     [A]----5km----[B]----3km----[C]
      |                         |
     2km                       4km
      |                         |
     [D]----6km----[E]----2km----[F]
      |             |
     3km           1km
      |             |
     [G]-----------[H]

Nodes: A,B,C,D,E,F,G,H (intersections)
Edges: (A,B)=5km, (B,C)=3km, (A,D)=2km, ...
Dijkstra/A* finds shortest path from A to F
```

---

## 7. Scaling

### Sharding
- **Tiles**: By (z, x, y) — each region can be separate cluster
- **Road graph**: By region (continent, country)
- **Places**: By region or by place_id hash

### Caching
- **CDN**: Tiles (90%+ hit rate)
- **Redis**: Routes (origin_cell, dest_cell) → route
- **Redis**: Geocode results
- **Redis**: Traffic segment speeds

### CDN
- Tile delivery at edge
- Reduce origin load by 10–100x

### Read Replicas
- PostgreSQL: Routing queries
- Elasticsearch: Place search

---

## 8. Failure Handling

### Tile Server Down
- CDN serves stale tiles
- Fallback to lower zoom or simplified tiles

### Routing Service Down
- Cache previous routes
- Degrade to straight-line ETA

### Traffic Data Stale
- Fallback to historical average
- Or speed limit only

### Data Center Failure
- Multi-region deployment
- Route by user region

---

## 9. Monitoring & Observability

### Key Metrics
| Metric | Target |
|--------|--------|
| Tile load p99 | < 200 ms |
| Route computation p99 | < 2 sec |
| Search p99 | < 500 ms |
| Cache hit rate (tiles) | > 90% |
| Traffic data freshness | < 5 min |

### Logging
- Slow route queries
- Geocode failures
- Tile generation errors

### Tracing
- End-to-end: Request → Tile/Route/Search → Response

---

## 10. Interview Tips

### Follow-up Questions
1. How would you add real-time road closures?
2. How do you handle routing in areas with poor map data?
3. How would you implement "avoid tolls" or "avoid highways"?
4. How do you scale the road graph for a continent?

### Common Mistakes
- Not considering tile pyramid and zoom levels
- Ignoring traffic in routing
- Overlooking turn restrictions
- Underestimating map data volume

### What to Emphasize
- Tile system (raster vs vector, pyramid)
- Routing (Dijkstra, A*, Contraction Hierarchies)
- Traffic pipeline (GPS → map-matching → aggregation)
- Geospatial indexing (R-tree, quadtree, Elasticsearch geo)
- CDN for tiles

---

## Appendix A: Contraction Hierarchies (Detailed)

### Preprocessing
1. Assign "importance" to each node (e.g., betweenness centrality, road class)
2. Contract nodes in order of increasing importance
3. When contracting node v: for each pair (u, w) of neighbors, if shortest path u→v→w is optimal, add shortcut (u, w) with that weight
4. Store shortcuts; original graph + shortcuts = "contraction hierarchy"

### Query
1. Run Dijkstra from origin, only following "upward" edges (to higher-level nodes)
2. Run Dijkstra from destination, backward, only "upward"
3. Meet at high-level nodes; shortest path = min over meeting points
4. Unpack path (replace shortcuts with original edges)

### Complexity
- Preprocessing: O(n log n) or more
- Query: O(√n) or better in practice
- Enables sub-second continental routing

---

## Appendix B: Map-Matching Algorithm

### Problem
GPS points are noisy; need to snap to road network.

### Approach
1. **Point-to-curve**: For each point, find nearest road segment
2. **HMM (Hidden Markov Model)**: States = road segments; emissions = GPS points; transition = road connectivity; find most likely path
3. **Incremental**: Process points in order; extend path greedily

### Output
- Sequence of (edge_id, offset) — which segment, how far along

---

## Appendix C: Vector Tile Format (Mapbox Vector Tiles)

### Structure
- Protocol Buffers
- Layers: points, lines, polygons
- Each layer: list of features with geometry + attributes
- Tiles are self-contained (no external refs)

### Advantages
- Small size (2–10 KB vs 20–50 KB raster)
- Client can style (colors, labels)
- One tile set for all zoom levels (client zooms)
- Accessibility (text layers)

---

## Appendix D: Transit Routing

### Graph Model
- **Stops**: Nodes
- **Connections**: Edge from (stop A, time t) to (stop B, time t') if trip connects them
- **Walking**: Edges between nearby stops
- **Transfer**: Walk between platforms

### Algorithm
- Dijkstra with time-dependent edges
- Or: RAPTOR (Round-Based Public Transit Optimized Router) — round-based, Pareto-optimal arrival times

### Data
- GTFS (General Transit Feed Specification): stops, routes, trips, stop_times
- Parse GTFS → build graph

---

## Appendix E: Geocoding Pipeline

### Address Parsing
- Input: "123 Main St, San Francisco, CA 94102"
- Tokenize: number, street, city, state, zip
- Normalize: "St" → "Street", "CA" → "California"
- Query: Elasticsearch with analyzed fields

### Ranking
- Exact match > partial match
- Distance from user location (if provided)
- Population (prefer larger city for ambiguity)

### Reverse Geocoding
- Point-in-polygon: Administrative boundaries (country, state, city)
- Nearest road: Snap to segment; interpolate address
- Combine: "123 Main St, San Francisco, CA, USA"

---

## Appendix F: Traffic Data Ingestion

### GPS Probe Pipeline
```
Phones → Batch (5 min) → Kafka → Map-Matching Service → Aggregation Service → Traffic Store
```

### Map-Matching
- Snap (lat, lng, speed) to road segment
- Output: (segment_id, speed, timestamp)
- Handle noise: median over multiple probes

### Aggregation
- Per segment, 5-min bucket: median speed, count
- Store: segment_id, bucket_start, avg_speed, sample_count
- Serve: Latest bucket + historical for prediction

---

## Appendix G: Place Index Schema (Elasticsearch)

```json
{
  "mappings": {
    "properties": {
      "name": {"type": "text", "analyzer": "standard"},
      "name_suggest": {"type": "completion", "contexts": [{"type": "geo", "precision": "1km"}]},
      "type": {"type": "keyword"},
      "location": {"type": "geo_point"},
      "address": {"type": "text"},
      "rating": {"type": "float"}
    }
  }
}
```

### Autocomplete Query
- Use `completion` suggester with geo context
- Prefix match on name
- Boost by distance from user

---

## Appendix H: Tile Generation Pipeline

### Input
- Road network (PostGIS)
- POI data
- Land use, water bodies
- Labels (placenames)

### Process
- For each zoom level (0–18):
  - Simplify geometry (Douglas-Peucker) based on zoom
  - Clip to tile bounds
  - Encode to vector tile format (MVT)
  - Write to storage (e.g., S3)

### Tools
- Tippecanoe (Mapbox)
- OpenMapTiles
- Custom (PostGIS ST_AsMVT)

---

## Appendix I: A* Heuristic

### Admissible Heuristic
- h(n) = straight-line distance from n to goal / max_speed
- Never overestimate; ensures optimal path
- For road networks: Euclidean distance / 120 km/h (highway speed)

### Why A* Faster Than Dijkstra
- Dijkstra explores in all directions
- A* prioritizes toward goal
- Fewer nodes explored
- Same result if heuristic admissible

---

## Appendix J: Turn Restriction Storage

### Format
- (from_edge_id, via_node_id, to_edge_id) = forbidden
- Example: No left turn from Main St onto Oak St

### Query
- At via_node, with outgoing edge to_edge
- Check: Is (incoming_edge, via_node, to_edge) forbidden?
- If yes: Exclude to_edge from neighbors in Dijkstra

---

## Appendix K: Raster vs Vector Tile Comparison

| Aspect | Raster | Vector |
|--------|--------|--------|
| File size | 20–50 KB | 2–10 KB |
| Styling | Server-side only | Client-side |
| Zoom | Pre-rendered per level | Single set, client zooms |
| Labels | Baked in | Separate layer, collidable |
| Accessibility | Limited | Text layers |
| Production | GDAL, Mapnik | Tippecanoe, Mapbox |
