# Part 6: Add Retry Logic for Google Calendar API

## Goal

Improve the resilience of the Google Calendar integration by implementing retry logic with exponential backoff to handle transient errors (e.g., rate limits, network glitches).

## Sub-tasks

1. **Design**: Create a `RetryTransport` that implements `http.RoundTripper`.
2. **Logic**: Handle specific HTTP status codes (429, 500, 502, 503, 504) and network errors.
3. **Backoff**: Implement exponential backoff with jitter.
4. **Integration**: Wrap the Google Calendar client's transport.
5. **Testing**: Verify retry behavior with unit tests.

---

## 6.1 Solution Status

### **Completed**

> [!NOTE]
> **Status**: Restored and Verified.
> **Verification**:
>
> - `internal/adapters/googlecalendar/retry.go` implemented.
> - `internal/adapters/googlecalendar/retry_test.go` passed.
> - Client integration in `client.go` verified.

## 6.2 Implementation Details

- **RetryTransport**: Custom struct wrapping the base transport.
- **Max Retries**: Default 3.
- **Backoff Strategy**:
  - Base Delay: 500ms
  - Max Delay: 5s
  - Multiplier: 2.0
  - Jitter: +/- 20%
- **Trigger Conditions**:
  - `429 Too Many Requests`
  - `500 Internal Server Error`
  - `502 Bad Gateway`
  - `503 Service Unavailable`
  - `504 Gateway Timeout`
  - Network errors (e.g., connection reset)
