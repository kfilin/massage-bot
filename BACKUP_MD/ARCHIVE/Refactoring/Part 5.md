# Part 5: Cache Free‑Busy Results (short‑term)

## Goal

Implement a short-term cache for free-busy results to reduce the number of calls to the Google Calendar API, improving performance and avoiding rate limits.

## Sub-tasks

1. **Analyze & Critique**: Evaluate the proposed caching strategy in `internal/services/appointment/service.go`.
2. **Setup Metrics**: Add `FreeBusyCacheHits` and `FreeBusyCacheMisses` to the monitoring package.
3. **Implement Caching**: Add the cache structure and logic to the `appointment.Service`.
4. **Testing**: Verify caching logic with unit tests and mocks.

---

## 5.1 Solution Status

### **Completed**

> [!NOTE]
> **Status**: Restored and Verified.
> **Verification**:
>
> - Metrics added to `internal/monitoring/metrics.go`.
> - Caching implemented in `internal/services/appointment/service.go`.
> - Verified with `internal/services/appointment/cache_test.go`.

## 5.2 Analyze & Critique Solution

The initial proposal was solid but required a few refinements implemented during development:

- **Caching Raw Data**: We chose to cache the raw `[]domain.TimeSlot` (busy intervals) instead of "available slots". This allows the service to correctly calculate availability for different service durations using the same cached data.
- **Cache Invalidation**: Crucially, we implemented automatic cache invalidation on both `CreateAppointment` and `CancelAppointment`. This prevents "double booking" that could occur if a user saw stale availability.
- **Shared Method**: A private `getFreeBusy` method was introduced to encapsulate the caching logic. It is used by both `GetAvailableTimeSlots` (for browsing) and `CreateAppointment` (for final validation).

## 5.2 Setup Metrics Solution

Two new Prometheus counters were added to `internal/monitoring/metrics.go`:

- `vera_freebusy_cache_hits_total`: Incremented whenever a request is served from the in-memory cache.
- `vera_freebusy_cache_misses_total`: Incremented when an API call to Google Calendar is required.

## 5.3 Implement Caching Solution

- **Struct**: Added `fbCache map[string]freeBusyEntry` and `fbCacheMu sync.RWMutex` to the `appointment.Service`.
- **Logic**: The `getFreeBusy` method checks the cache using a key composed of `calendarID:date`. If a non-expired entry is found, it's returned immediately.
- **TTL**: The cache duration is controlled by `config.Default.CacheTTL` (default 2 minutes).
- **Cleanup**: Removed unused legacy cache fields from the `Service` struct.

## 5.4 Testing Solution

A new test file `internal/services/appointment/cache_test.go` was created. It verifies:

- Repo is only called once for multiple consecutive `GetAvailableTimeSlots` calls.
- `CreateAppointment` hits the cache for validation but invalidates it for subsequent calls.
- `CancelAppointment` correctly invalidates the cache.
- Expired entries trigger a fresh repo call.
