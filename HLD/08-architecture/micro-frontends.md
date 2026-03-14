# Micro Frontends

## 1. Concept Overview

### Definition
**Micro frontends** extend the microservices concept to the frontend. Instead of a single monolithic frontend application, the UI is split into independent pieces owned by different teams. Each team develops, deploys, and maintains their own frontend module independently, while these modules compose together at runtime (or build-time) to form a cohesive user experience.

### Purpose
- **Independent deployment**: Deploy frontend changes without coordinating with other teams
- **Team autonomy**: Each team owns their vertical slice (backend + frontend)
- **Technology diversity**: Different teams can use React, Vue, Angular, or Svelte
- **Scalability**: Large teams can work in parallel without merge conflicts
- **Incremental migration**: Migrate from monolith to micro frontends gradually

### Problems It Solves
- **Monolithic frontend bottleneck**: Single codebase, single deployment, merge conflicts
- **Release cycle coupling**: One team's change blocks entire release
- **Technology lock-in**: Entire app tied to one framework
- **Team scaling**: Large frontend team creates coordination overhead
- **Legacy migration**: Can't rewrite entire app; need incremental approach

---

## 2. Real-World Motivation

### IKEA
- **Scale**: 4,000+ developers across 100+ teams
- **Approach**: Micro frontends per product area (e.g., product catalog, cart, checkout)
- **Benefits**: Teams deploy independently; no "big bang" releases
- **Integration**: Module federation, shared design system

### Spotify
- **Web player**: Micro frontends for different sections (browse, search, library, playlist)
- **Team structure**: Each squad owns a vertical (e.g., Search squad owns search UI + backend)
- **Technology**: Mix of React, legacy jQuery; gradual migration

### Zalando
- **E-commerce**: Product page, cart, checkout, account as separate micro frontends
- **Project Mosaic**: Open-source framework for micro frontends
- **Deployment**: Each team deploys their fragment independently

### DAZN (Streaming)
- **Sports streaming**: Live streaming, VOD, user profile as separate apps
- **Single-spa**: Runtime composition framework
- **Deployment**: Independent CI/CD per micro frontend

### American Express
- **Enterprise scale**: Multiple teams, multiple products
- **Composition**: Server-side composition (SSI/ESI) for some pages
- **Design system**: Shared component library across micro frontends

---

## 3. Architecture Diagrams

### Monolith vs Micro Frontends

```
MONOLITHIC FRONTEND
┌─────────────────────────────────────────────────────────────────┐
│                    SINGLE FRONTEND APPLICATION                    │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐    │
│  │ Header  │ │ Product │ │  Cart   │ │Checkout │ │ Account │    │
│  │  Team   │ │  Team   │ │  Team   │ │  Team   │ │  Team   │    │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘    │
│       │           │           │           │           │         │
│       └───────────┴───────────┴───────────┴───────────┘         │
│                              │                                   │
│                    ONE CODEBASE, ONE DEPLOYMENT                   │
│                    Merge conflicts, shared releases               │
└─────────────────────────────────────────────────────────────────┘

MICRO FRONTENDS
┌─────────────────────────────────────────────────────────────────┐
│                    COMPOSITION LAYER (Shell/Container)            │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │  Route: /product → Product MF | /cart → Cart MF | /checkout  │ │
│  └─────────────────────────────────────────────────────────────┘ │
│              │                    │                    │         │
│   ┌─────────────┐      ┌─────────────┐      ┌─────────────┐      │
│   │  Product    │      │    Cart     │      │  Checkout   │      │
│   │  Micro      │      │    Micro    │      │   Micro     │      │
│   │  Frontend   │      │  Frontend   │      │  Frontend   │      │
│   │  (Team A)   │      │  (Team B)   │      │  (Team C)   │      │
│   │  React      │      │  Vue        │      │  Angular    │      │
│   └─────────────┘      └─────────────┘      └─────────────┘      │
│   Independent deploy │ Independent deploy │ Independent deploy   │
└─────────────────────────────────────────────────────────────────┘
```

### Composition Approaches

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    MICRO FRONTEND COMPOSITION APPROACHES                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  1. BUILD-TIME COMPOSITION (NPM packages)                                │
│     ┌──────────┐  ┌──────────┐  ┌──────────┐                             │
│     │  MF A    │  │  MF B    │  │  MF C    │  → npm install → Shell       │
│     │  (pkg)   │  │  (pkg)   │  │  (pkg)   │     Single bundle            │
│     └──────────┘  └──────────┘  └──────────┘     Coupled deployment        │
│                                                                          │
│  2. RUNTIME COMPOSITION (Module Federation)                               │
│     ┌──────────┐  ┌──────────┐  ┌──────────┐                             │
│     │  MF A    │  │  MF B    │  │  MF C    │  → Shell loads at runtime    │
│     │  (host)  │  │  (remote)│  │  (remote)│     Independent deploys      │
│     └──────────┘  └──────────┘  └──────────┘     Shared chunks            │
│                                                                          │
│  3. IFRAME COMPOSITION                                                    │
│     ┌─────────────────────────────────────────────────────────────────┐  │
│     │  Shell                                                           │  │
│     │  ┌─────────────────────────────────────────────────────────────┐ │  │
│     │  │  <iframe src="https://cart.example.com/embed">              │ │  │
│     │  │  <iframe src="https://product.example.com/embed">         │ │  │
│     │  └─────────────────────────────────────────────────────────────┘ │  │
│     └─────────────────────────────────────────────────────────────────┘  │
│     Isolation: Strong | Communication: postMessage | Styling: Isolated   │
│                                                                          │
│  4. WEB COMPONENTS (Custom Elements)                                      │
│     <product-catalog></product-catalog>                                   │
│     <shopping-cart></shopping-cart>                                       │
│     Framework-agnostic; each MF loads its own script                     │
│                                                                          │
│  5. SERVER-SIDE COMPOSITION (SSI / ESI)                                   │
│     <html>                                                                │
│       <!--#include virtual="/header" -->                                  │
│       <!--#include virtual="/product/123" -->                              │
│       <!--#include virtual="/footer" -->                                  │
│     </html>                                                               │
│     Server assembles HTML; each fragment from different service           │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Module Federation Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    WEBPACK MODULE FEDERATION                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   SHELL (Host)                    REMOTE APPS                             │
│   ┌─────────────────┐             ┌─────────────────┐                      │
│   │  Loads remotes  │             │  Product MF     │                      │
│   │  at runtime     │◄────────────│  exposes:       │                      │
│   │                 │   fetch     │  - ProductPage  │                      │
│   │  shared: React  │             │  - ProductCard  │                      │
│   └────────┬────────┘             └─────────────────┘                      │
│            │                      ┌─────────────────┐                      │
│            │◄─────────────────────│  Cart MF        │                      │
│            │   fetch              │  exposes:       │                      │
│            │                      │  - CartWidget   │                      │
│            │                      └─────────────────┘                      │
│            │                                                                 │
│   Runtime: Shell loads ProductPage, CartWidget from different origins      │
│   Shared chunk: React loaded once, shared across all MFs                   │
│   Deployment: Each MF deployed independently; no shell rebuild needed      │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Integration Strategies

| Strategy | Build-time | Runtime | Isolation | Complexity |
|----------|------------|---------|-----------|------------|
| **NPM packages** | Yes | No | Low | Low |
| **Module Federation** | No | Yes | Medium | Medium |
| **iframe** | No | Yes | High | Low |
| **Web Components** | No | Yes | Medium | Medium |
| **Server-Side (SSI/ESI)** | No | Yes | High | Medium |

### Build-Time vs Runtime Integration

**Build-Time**
- Shell imports MFs as packages; single bundle
- Pros: Simple, shared dependencies, no runtime loading
- Cons: Coupled deployment; shell rebuilds when any MF changes
- Use when: Small teams, infrequent releases

**Runtime**
- Shell loads MFs at runtime (fetch from CDN, different URLs)
- Pros: Independent deployment; deploy MF without touching shell
- Cons: Loading complexity, version compatibility, shared chunks
- Use when: Large teams, independent release cycles

### Shared Dependencies
- **Shared**: React, design system, shared UI components
- **Version alignment**: All MFs should use compatible versions
- **Module Federation**: Single shared chunk; loaded once
- **Design system**: NPM package or shared remote

### Routing
- **Shell routing**: Shell owns route; loads appropriate MF per route
- **Route-based splitting**: `/product/*` → Product MF, `/cart` → Cart MF
- **Nested routes**: MF can have nested routes within its scope

---

## 5. Numbers

| Metric | Monolith | Micro Frontends |
|--------|----------|-----------------|
| Deployment frequency | Weekly | Per-team daily |
| Bundle size (initial) | Single | Shell + lazy MF |
| Load time (first visit) | 1-3s | 1-4s (MF loading) |
| Team size | 10-50+ | 2-10 per MF |
| Merge conflicts | High | Low (separate repos) |
| Shared chunk overhead | N/A | 50-200KB (React, etc.) |

### IKEA Scale
- 100+ teams, 4,000+ developers
- 100+ micro frontends
- Multiple deployments per day

### Bundle Size Considerations
- **Shell**: 50-100KB (routing, auth)
- **Each MF**: 100-500KB (lazy loaded)
- **Shared**: 100-200KB (React, design system)
- **Total**: Similar to monolith if lazy loading; risk of duplication without Module Federation

---

## 6. Tradeoffs

### Micro Frontends Tradeoffs

| Benefit | Cost |
|---------|------|
| Independent deployment | Bundle size duplication, consistency challenges |
| Team autonomy | Shared state complexity, design system consistency |
| Technology diversity | Integration complexity, version conflicts |
| Incremental migration | Operational overhead |

### Build-Time vs Runtime

| Aspect | Build-Time | Runtime |
|--------|------------|---------|
| **Deployment** | Coupled | Independent |
| **Bundle** | Single | Multiple (lazy) |
| **Complexity** | Lower | Higher |
| **Flexibility** | Lower | Higher |
| **Use case** | Small teams | Large orgs |

### Consistency vs Autonomy

| More Autonomy | More Consistency |
|---------------|-------------------|
| Each team owns UI completely | Shared design system |
| Different tech stacks | Single framework |
| Independent styling | Centralized design tokens |
| Risk: Inconsistent UX | Risk: Bottleneck |

---

## 7. Variants / Implementations

### Frameworks

**Single-SPA**
- Meta-framework; orchestrates multiple frameworks (React, Vue, Angular)
- Each MF is a "parcel" or "application"
- Routing: Single-SPA routes to appropriate MF

**Module Federation (Webpack 5)**
- Native Webpack; no meta-framework
- Exposes/remotes; shared chunks
- Used by: IKEA, many enterprises

**iframe**
- Simple; strong isolation
- postMessage for communication
- Styling isolated; UX limitations (height, scroll)

**Web Components**
- Custom elements; framework-agnostic
- Each MF wraps its app in custom element
- Good for gradual migration

**Server-Side (SSI/ESI)**
- Edge Side Includes (ESI): CDN assembles fragments
- Server Side Includes (SSI): Origin assembles
- Each fragment from different service
- No shared state; each fragment independent

### Implementations
- **IKEA**: Module Federation, shared design system
- **Spotify**: Single-SPA, React + legacy
- **Zalando**: Project Mosaic (SSI/ESI)
- **DAZN**: Single-SPA
- **Amazon**: Some pages use SSI for fragments

---

## 8. Scaling Strategies

### Team Scaling
- Add new MF for new product area
- New team owns new MF
- Shell updated to route to new MF

### Performance
- **Lazy loading**: Load MF only when route matches
- **Preloading**: Preload MF on hover or when route is likely
- **Code splitting**: Each MF is a chunk; load on demand
- **CDN**: Each MF deployed to CDN; cache independently

### Shared State
- **Props/callbacks**: Shell passes to MF (simple)
- **Event bus**: Custom events for cross-MF communication
- **Shared state library**: Redux, Zustand with shared store
- **URL/query params**: State in URL for shareability

---

## 9. Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| MF fails to load | Blank section or error boundary | Fallback UI, retry, offline message |
| MF returns 404 | Broken page | Version pinning, CDN fallback |
| Version mismatch | Runtime errors | Shared dependency versioning |
| Shared chunk conflict | Multiple React instances | Module Federation shared scope |
| Styling conflict | UI broken | CSS isolation (shadow DOM, BEM) |
| Slow MF | Blocking render | Lazy load, skeleton, timeout |

### Mitigation Strategies
- **Error boundaries**: Catch MF errors; show fallback
- **Version pinning**: Shell loads specific MF version
- **Health checks**: Monitor MF availability
- **Graceful degradation**: Hide MF if load fails; show message

---

## 10. Performance Considerations

- **Initial load**: Shell + first MF; keep shell small
- **Lazy loading**: Load MFs on route change; don't load all upfront
- **Shared chunks**: Avoid duplicate React, etc.; use Module Federation
- **Bundle size**: Each MF can bloat; enforce size budgets
- **Caching**: Each MF has own cache key; version in URL
- **Prefetching**: Prefetch next likely MF

---

## 11. Use Cases

| Use Case | Why Micro Frontends |
|----------|---------------------|
| **Large e-commerce** (IKEA, Zalando) | Many teams, independent product areas |
| **Enterprise SaaS** | Multiple teams, different modules |
| **Legacy migration** | Incremental rewrite; new MF alongside old |
| **Multi-brand** | Same shell, different MFs per brand |
| **B2B platforms** | Different teams own different features |
| **Streaming platforms** (Spotify, DAZN) | Browse, search, player as separate MFs |

---

## 12. Comparison Tables

### Composition Approach Comparison

| Approach | Isolation | Complexity | Independent Deploy | Best For |
|----------|-----------|------------|-------------------|----------|
| **NPM packages** | Low | Low | No | Small teams |
| **Module Federation** | Medium | Medium | Yes | React/Webpack orgs |
| **Single-SPA** | Medium | Medium | Yes | Multi-framework |
| **iframe** | High | Low | Yes | Strong isolation |
| **Web Components** | Medium | Medium | Yes | Framework-agnostic |
| **SSI/ESI** | High | Medium | Yes | Server-rendered |

### When to Use Micro Frontends

| Use Micro Frontends | Use Monolith |
|---------------------|--------------|
| 5+ frontend teams | 1-2 frontend devs |
| Independent release cycles | Coordinated releases OK |
| Different tech stacks | Single framework |
| Legacy migration | Greenfield |
| Large org (100+ devs) | Small org |

---

## 13. Code or Pseudocode

### Module Federation Config (Webpack)

```javascript
// shell/webpack.config.js (Host)
module.exports = {
  plugins: [
    new ModuleFederationPlugin({
      name: 'shell',
      remotes: {
        product: 'product@https://cdn.example.com/product/remoteEntry.js',
        cart: 'cart@https://cdn.example.com/cart/remoteEntry.js',
      },
      shared: {
        react: { singleton: true, requiredVersion: '^18.0.0' },
        'react-dom': { singleton: true, requiredVersion: '^18.0.0' },
      },
    }),
  ],
};

// product/webpack.config.js (Remote)
module.exports = {
  plugins: [
    new ModuleFederationPlugin({
      name: 'product',
      filename: 'remoteEntry.js',
      exposes: {
        './ProductPage': './src/ProductPage',
        './ProductCard': './src/ProductCard',
      },
      shared: {
        react: { singleton: true },
        'react-dom': { singleton: true },
      },
    }),
  ],
};
```

### Shell Loading MF (React)

```javascript
// Shell App
import React, { lazy, Suspense } from 'react';

const ProductPage = lazy(() => import('product/ProductPage'));
const CartWidget = lazy(() => import('cart/CartWidget'));

function App() {
  return (
    <div>
      <Header />
      <Suspense fallback={<Loading />}>
        <Routes>
          <Route path="/product/*" element={<ProductPage />} />
          <Route path="/cart" element={<CartWidget />} />
        </Routes>
      </Suspense>
    </div>
  );
}
```

### Cross-MF Communication (Event Bus)

```javascript
// Shared event bus
const eventBus = {
  subscribe(event, callback) {
    window.addEventListener(event, callback);
  },
  publish(event, data) {
    window.dispatchEvent(new CustomEvent(event, { detail: data }));
  },
};

// Cart MF: publish when item added
eventBus.publish('cart:itemAdded', { productId: 123, quantity: 1 });

// Header MF: subscribe to update badge
eventBus.subscribe('cart:itemAdded', () => {
  updateCartBadge();
});
```

---

## 14. Interview Discussion

### Key Points to Articulate
1. **Definition**: Micro frontends = microservices for frontend; independent teams own UI pieces
2. **Composition**: Build-time (NPM) vs runtime (Module Federation, iframe, web components)
3. **Tradeoffs**: Bundle size, consistency, shared state complexity
4. **When**: Large teams, independent releases, legacy migration
5. **Examples**: IKEA, Spotify, Zalando

### Common Interview Questions
- **"How would you split a monolith frontend?"** → By route/feature; each team owns vertical; use Module Federation or Single-SPA
- **"Build-time vs runtime composition?"** → Build-time: simpler, coupled deploy. Runtime: independent deploy, more complex
- **"How do you handle shared state?"** → Event bus, URL state, shared Redux; or minimize cross-MF state
- **"What about bundle size?"** → Shared chunks (Module Federation), lazy loading, design system as shared
- **"When would you NOT use micro frontends?"** → Small team, simple app, need strong consistency

### Red Flags to Avoid
- Suggesting micro frontends for 2-person team
- Ignoring bundle size and consistency
- Not considering shared state
- Over-complicating with iframes when Module Federation suffices

### Ideal Answer Structure
1. Define micro frontends (extend microservices to frontend)
2. Explain composition approaches (Module Federation, iframe, etc.)
3. Discuss tradeoffs (bundle size, consistency, shared state)
4. Give real examples (IKEA, Spotify)
5. When to use vs not use
