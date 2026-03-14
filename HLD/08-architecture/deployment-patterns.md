# Deployment Patterns

## 1. Concept Overview

### Definition
**Deployment patterns** are strategies for releasing new versions of software to production with minimal risk and downtime. They control how traffic is shifted from old to new versions, how rollbacks are performed, and how to validate changes before full rollout.

### Purpose
- **Zero/minimal downtime**: Release without taking the system offline
- **Risk reduction**: Validate new version with limited exposure before full rollout
- **Fast rollback**: Quickly revert if issues are detected
- **Gradual validation**: Canary, A/B testing to measure impact
- **Feature control**: Feature flags for runtime toggling

### Problems It Solves
- **Big bang deployments**: High risk; one bad deploy takes down everything
- **Long rollback time**: Need to revert quickly when issues arise
- **Unvalidated changes**: Deploy to all users without testing in production
- **Feature coupling**: Can't release partially; all-or-nothing

---

## 2. Real-World Motivation

### Netflix
- **Canary deployment**: Deploy to small subset of users; monitor error rates, latency
- **Spinnaker**: Open-source CD platform; supports canary, red/black (blue-green)
- **Chaos Engineering**: Validate resilience during deployment
- **Scale**: 200M+ subscribers; deployments multiple times per day

### Facebook
- **Dark launches**: Deploy code to production but don't enable for users; validate infrastructure
- **Gatekeeper**: Feature flag system; control rollout by percentage, user segment
- **Gradual rollout**: 1% → 5% → 20% → 50% → 100% over hours/days

### Amazon
- **Blue-green**: Maintain two identical environments; switch traffic
- **Deployment frequency**: Every 11.7 seconds (2014); even higher now
- **Feature flags**: Launch Dark (internal tool) for gradual feature rollout

### Google
- **Canary**: Deploy to single server/canary; compare metrics; expand
- **Progressive rollout**: Staged rollout across regions
- **Borg**: Infrastructure for deployment at scale

### Uber
- **uDeploy**: Canary, blue-green, rolling
- **Feature flags**: LaunchDarkly, custom; control by city, user segment
- **Shadow traffic**: Send copy of traffic to new version; validate without user impact

---

## 3. Architecture Diagrams

### Blue-Green Deployment

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     BLUE-GREEN DEPLOYMENT                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   LOAD BALANCER                                                          │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  Traffic: 100% to ACTIVE environment                             │   │
│   └───────────────────────────┬─────────────────────────────────────┘   │
│                                │                                          │
│                    ┌───────────▼───────────┐                              │
│                    │  BLUE (Active) v1.0  │  ◄── All traffic            │
│                    │  ┌───┐ ┌───┐ ┌───┐   │                              │
│                    │  │ A │ │ A │ │ A │   │                              │
│                    │  └───┘ └───┘ └───┘   │                              │
│                    └──────────────────────┘                              │
│                                                                          │
│                    ┌──────────────────────┐                              │
│                    │  GREEN (Idle) v1.1   │  ◄── Deploy new version      │
│                    │  ┌───┐ ┌───┐ ┌───┐   │     No traffic yet           │
│                    │  │ N │ │ N │ │ N │   │                              │
│                    │  └───┘ └───┘ └───┘   │                              │
│                    └──────────────────────┘                              │
│                                                                          │
│   SWITCH: Update LB to point to GREEN → Instant cutover                   │
│   ROLLBACK: Point LB back to BLUE → Instant revert                       │
│   DOWNTIME: None (if DB schema compatible)                                │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Canary Deployment

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     CANARY DEPLOYMENT                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   LOAD BALANCER (weighted / subset routing)                              │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  90% → Stable (v1.0)    10% → Canary (v1.1)                     │   │
│   └───────────────────────────┬───────────────────┬─────────────────┘   │
│                               │                   │                      │
│                               ▼                   ▼                      │
│                    ┌──────────────────┐  ┌──────────────────┐            │
│                    │  STABLE v1.0     │  │  CANARY v1.1     │            │
│                    │  ┌───┐ ┌───┐     │  │  ┌───┐            │            │
│                    │  │ A │ │ A │ ... │  │  │ N │  (1-2     │            │
│                    │  └───┘ └───┘     │  │  └───┘   inst)   │            │
│                    └──────────────────┘  └──────────────────┘            │
│                                                                          │
│   Monitor: Error rate, latency, business metrics                          │
│   If OK: Increase canary % (10% → 25% → 50% → 100%)                       │
│   If BAD: Rollback canary; traffic 100% to stable                         │
│                                                                          │
│   Netflix: Canary by user %; compare canary vs baseline metrics           │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Rolling Update

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     ROLLING UPDATE                                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   BEFORE:  [v1] [v1] [v1] [v1] [v1]  (5 instances)                       │
│                                                                          │
│   STEP 1: [v1] [v1] [v1] [v1] [v2]  Replace 1 instance                  │
│   STEP 2: [v1] [v1] [v1] [v2] [v2]  Replace another                     │
│   STEP 3: [v1] [v1] [v2] [v2] [v2]  ...                                 │
│   STEP 4: [v1] [v2] [v2] [v2] [v2]  ...                                 │
│   STEP 5: [v2] [v2] [v2] [v2] [v2]  All on v2                           │
│                                                                          │
│   - No separate "green" environment; replace in place                     │
│   - Configurable: maxSurge, maxUnavailable (Kubernetes)                  │
│   - Slower than blue-green; lower resource cost                          │
│   - Both versions run simultaneously during rollout                      │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Feature Flags

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     FEATURE FLAGS                                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Application Code                                                        │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  if (featureFlags.isEnabled("new_checkout", user_id)) {           │   │
│   │    renderNewCheckout();                                          │   │
│   │  } else {                                                        │   │
│   │    renderOldCheckout();                                           │   │
│   │  }                                                                │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                │                                         │
│                                ▼                                         │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  Feature Flag Service (LaunchDarkly, Unleash, custom)            │   │
│   │  - Rollout: 10% of users                                        │   │
│   │  - Targeting: user_id in [list], country=US                       │   │
│   │  - Kill switch: disable without redeploy                         │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│   Deploy code with both paths; toggle at runtime                          │
│   No redeploy to enable/disable feature                                  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### A/B Testing

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     A/B TESTING                                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Users randomly assigned to variant:                                    │
│                                                                          │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                  │
│   │  Control A  │    │ Variant B   │    │ Variant C   │                  │
│   │  50% users  │    │  25% users  │    │  25% users  │                  │
│   │  Old UI     │    │  New UI v1  │    │  New UI v2  │                  │
│   └─────────────┘    └─────────────┘    └─────────────┘                  │
│          │                   │                   │                       │
│          └───────────────────┴───────────────────┘                       │
│                              │                                            │
│                    Metrics: conversion, engagement, revenue                │
│                    Statistical significance before rollout                │
│                                                                          │
│   Same deployment; different experience by user segment                   │
│   Used for: UI experiments, pricing, algorithms                           │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Shadow / Dark Launch

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     SHADOW / DARK LAUNCH                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   PRODUCTION TRAFFIC                                                      │
│   ┌─────────────────────────────────────────────────────────────────┐   │
│   │  Request ──► [Load Balancer] ──► Stable v1.0 (response to user) │   │
│   │                     │                                            │   │
│   │                     │ COPY of request (async)                     │   │
│   │                     ▼                                            │   │
│   │              Shadow v1.1 (no response to user)                  │   │
│   │              - Validate new version handles real traffic          │   │
│   │              - Compare latency, errors                            │   │
│   │              - User never sees shadow response                    │   │
│   └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│   DARK LAUNCH: Code deployed, feature OFF                                │
│   - Validate infrastructure, DB migrations                               │
│   - Facebook: Deploy code, enable for 0% initially                        │   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Core Mechanics

### Blue-Green
- Two identical environments (Blue, Green)
- Deploy new version to idle environment
- Switch traffic (load balancer, DNS) to new environment
- Rollback: Switch back
- **Database**: Must handle schema; often run migrations before switch

### Canary
- New version receives small % of traffic (e.g., 5-10%)
- Monitor metrics; compare canary vs baseline
- Gradually increase if healthy; rollback if not
- **Selection**: Random %, user segment, or specific instances

### Rolling Update
- Replace instances one-by-one (or in batches)
- No second environment; in-place replacement
- Kubernetes: `Deployment` with `strategy: RollingUpdate`
- `maxUnavailable: 0` = zero downtime (surge new before killing old)

### Feature Flags
- Runtime configuration; no redeploy to toggle
- Targeting: user ID, %, country, cohort
- Kill switch: Disable feature instantly
- Stale flags: Remove after feature is permanent

### A/B Testing
- Split traffic by variant; measure outcome
- Requires statistical rigor (sample size, significance)
- Often implemented via feature flags with % split

### Shadow Launch
- Send copy of traffic to new version; don't use response
- Validates new version under real load
- Dark launch: Deploy code, don't enable; validate infra

---

## 5. Numbers

| Pattern | Downtime | Rollback Time | Resource Cost | Complexity |
|---------|----------|---------------|---------------|------------|
| **Blue-Green** | Zero | Instant (switch) | 2x during deploy | Low |
| **Canary** | Zero | Minutes (drain canary) | 1.x (canary instances) | Medium |
| **Rolling** | Zero (if maxUnavail=0) | Minutes (reverse rollout) | 1x | Low |
| **Feature flags** | Zero | Instant (toggle) | 1x | Low |
| **Shadow** | Zero | N/A (no user traffic) | 1.x (shadow instances) | High |

### Netflix Canary
- Canary: 1% of users initially
- Metrics: Error rate, latency, playback success
- Rollout: 1% → 5% → 25% → 50% → 100% over hours
- Automated: Spinnaker automates canary analysis

### Facebook Dark Launch
- Deploy to 100% of servers; feature enabled for 0%
- Validate: No crashes, infra stable
- Then: Enable for 1% → 5% → ... via Gatekeeper

---

## 6. Tradeoffs

### Pattern Selection

| Need | Pattern |
|------|---------|
| Fast rollback | Blue-Green |
| Validate with real traffic | Canary |
| Minimal resources | Rolling |
| Runtime control | Feature flags |
| Experiment | A/B testing |
| Validate without user impact | Shadow |

### Blue-Green vs Canary

| Aspect | Blue-Green | Canary |
|--------|------------|--------|
| **Resources** | 2x during deploy | 1.x (small canary) |
| **Rollback** | Instant | Drain canary |
| **Validation** | All-or-nothing | Gradual |
| **Risk** | Higher (100% switch) | Lower (limited exposure) |

### Rolling vs Blue-Green

| Aspect | Rolling | Blue-Green |
|--------|---------|------------|
| **Resources** | 1x | 2x |
| **Rollback** | Reverse rollout | Instant switch |
| **Speed** | Slower | Instant |
| **Complexity** | Lower | Higher (two envs) |

---

## 7. Variants / Implementations

### Variants

**Red-Black (AWS)**
- Same as Blue-Green; AWS terminology
- Elastic Beanstalk, CodeDeploy support

**Phased Rollout**
- Canary with multiple phases
- 1% → 10% → 50% → 100%
- Automated or manual gates

**Ring Deployment (Microsoft)**
- Ring 0: Canary (internal)
- Ring 1-N: Progressive production
- Each ring is a deployment stage

### Implementations
- **Kubernetes**: RollingUpdate, Blue-Green via two Deployments
- **Spinnaker**: Canary, red/black pipelines
- **Argo CD**: GitOps; rolling or blue-green
- **LaunchDarkly, Unleash**: Feature flags
- **Istio**: Traffic splitting for canary (weighted routing)

---

## 8. Scaling Strategies

- **Blue-Green**: Scale both environments; switch when ready
- **Canary**: Scale canary with traffic %; e.g., 10% traffic = 10% of instances
- **Rolling**: Scale during rollout; ensure capacity for maxSurge
- **Feature flags**: No scaling impact; runtime only

---

## 9. Failure Scenarios

| Scenario | Mitigation |
|----------|------------|
| **Bad deploy (Blue-Green)** | Switch back to old; instant rollback |
| **Canary shows errors** | Stop rollout; drain canary; fix |
| **Rolling: new version fails** | Pause rollout; rollback failed instances |
| **Feature flag misconfiguration** | Toggle off; fix config |
| **DB migration breaks** | Blue-Green: run migration before switch; have backward-compatible migrations |
| **Traffic spike during deploy** | Ensure capacity; canary limits blast radius |

---

## 10. Performance Considerations

- **Blue-Green**: 2x resources during deploy; plan capacity
- **Canary**: Small overhead; monitor canary latency
- **Feature flags**: Latency for flag lookup (cache at edge)
- **Shadow**: Double load on backend during shadow; ensure capacity

---

## 11. Use Cases

| Use Case | Pattern |
|----------|---------|
| **High-availability app** | Blue-Green, Rolling |
| **Validate new version** | Canary, Shadow |
| **Gradual feature release** | Feature flags, Canary |
| **Experiment (UI, algorithm)** | A/B testing |
| **Infra validation** | Dark launch |
| **Database migration** | Blue-Green (migrate before switch) |

---

## 12. Comparison Tables

### Deployment Pattern Matrix

| Pattern | Downtime | Rollback | Resources | Risk | Best For |
|---------|----------|----------|-----------|------|----------|
| **Blue-Green** | Zero | Instant | 2x | Medium | Critical apps |
| **Canary** | Zero | Fast | 1.x | Low | Large scale |
| **Rolling** | Zero | Medium | 1x | Low | General |
| **Feature flags** | Zero | Instant | 1x | Low | Feature control |
| **A/B test** | Zero | Toggle | 1x | Low | Experiments |
| **Shadow** | Zero | N/A | 1.x | None | Validation |

### When to Use Each

| Situation | Recommendation |
|-----------|----------------|
| Need instant rollback | Blue-Green |
| Large user base, validate first | Canary |
| Limited resources | Rolling |
| Release feature gradually | Feature flags |
| Test two UIs | A/B testing |
| Validate new service | Shadow |

---

## 13. Code or Pseudocode

### Kubernetes Rolling Update

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1        # Add 1 extra during rollout
      maxUnavailable: 0  # Zero downtime
  template:
    spec:
      containers:
      - name: myapp
        image: myapp:v1.1
```

### Istio Canary (Weighted Routing)

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: myapp
spec:
  hosts:
  - myapp
  http:
  - match:
    - headers:
        canary:
          exact: "true"
    route:
    - destination:
        host: myapp
        subset: v2
      weight: 100
  - route:
    - destination:
        host: myapp
        subset: v1
      weight: 90
    - destination:
        host: myapp
        subset: v2
      weight: 10
```

### Feature Flag (Pseudocode)

```python
def get_checkout_ui(user_id: str):
    if feature_flags.is_enabled("new_checkout", user_id):
        return render_new_checkout()
    return render_old_checkout()

# LaunchDarkly / Unleash style
# Rollout: 10% of users
# Targeting: user_id % 10 == 0
```

### Blue-Green Switch (Pseudocode)

```python
def switch_traffic(to_green: bool):
    if to_green:
        load_balancer.set_backend("green-pool")
        # Green now receives 100% traffic
    else:
        load_balancer.set_backend("blue-pool")
        # Rollback to blue
```

---

## 14. Interview Discussion

### Key Points to Articulate
1. **Blue-Green**: Two envs; instant switch; 2x resources
2. **Canary**: Small % to new version; monitor; gradual increase
3. **Rolling**: Replace instances in place; zero downtime with maxUnavailable=0
4. **Feature flags**: Runtime toggle; no redeploy; kill switch
5. **A/B testing**: Split traffic; measure; statistical significance
6. **Shadow**: Copy traffic to new version; validate without user impact

### Common Interview Questions
- **"How would you deploy with zero downtime?"** → Blue-Green or Rolling (maxUnavailable=0)
- **"How does Netflix do canary?"** → Spinnaker; 1% → 5% → ...; compare error rate, latency
- **"What's a dark launch?"** → Deploy code, feature off; validate infra
- **"Feature flags vs canary?"** → Flags: runtime toggle, no deploy. Canary: deploy new version, limit traffic
- **"When would you use blue-green vs rolling?"** → Blue-Green: need instant rollback. Rolling: want to save resources

### Red Flags to Avoid
- Suggesting downtime for deployment
- Not considering rollback
- Ignoring database migrations
- Over-complicating for simple apps

### Ideal Answer Structure
1. List patterns (Blue-Green, Canary, Rolling, Feature flags)
2. Explain tradeoffs (resources, rollback speed)
3. Give examples (Netflix canary, Facebook dark launch)
4. Recommend based on context (scale, resources, risk tolerance)
