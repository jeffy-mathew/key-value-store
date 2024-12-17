# Key-Value Store Improvement Notes

## Current State
- In-memory key-value store with HTTP API
- Basic Set, Get, and Delete operations
- Benchmark tests for API and store operations
- Integration tests
- Unit tests

## Design choices
I have implemented a simple in memory key value store in this repo. 

I have also tried adding a persitence layer, but dropped it due to time constraints. Ideally it must run at a configurable fixed interval so that the data can be flushed to an external storage which implements an expected interface. 

I have also added context arguments in all methods, but defered the implementation of it.

I have deviated from the requirement of having `{"key":"value"}` as the payload. Instead, I have used the key-value struct for efficient handling of the data and maintenance.
```
type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
``` 


Following are the scope for improvements that can be made into the key-value store that comes into my mind:

### Protocol & API Enhancements
- Implement gRPC with Protocol Buffers
  - Better performance and type safety
  - Code generation for multiple languages
- Add RESP (Redis Serialization Protocol) support
  - Redis client compatibility
  - Efficient binary serialization

### DX Improvements
- Serve OpenAPI documentation
  - Interactive API documentation
  - Client code generation

### Data Management
- **TTL & Eviction**
  - Add key expiration support
  - Implement eviction strategies (LRU, LFU)
  - Background cleanup of expired keys
- **Persistence**
  - Periodic snapshots
  - Graceful shutdown with state preservation

### Security & Reliability
- **Context Propagation**
  - Proper cancellation handling
  - Request timeouts
  - Graceful shutdown signals
  - Dependency cleanup


# Performance Optimization Notes

## Current Benchmark Areas
- HTTP API Operations (SET, GET, DELETE)
- Direct Store Operations

## Performance Improvements

### Memory Management
  - Use sync.Pool for []byte buffers

### Concurrency Optimizations
  - Replace global mutex with sharded locks

### Data Structure Optimizations
  - Use string interning for repeated keys
  - Implement key compression for similar prefixes
  - Implement value compression for large values

### 4. HTTP Optimizations
- **Request Processing**
  - Optimize JSON serialization/deserialization
  - Use Protocol Buffers for binary encoding

- **High Concurrency**
  - Implement backpressure mechanisms
  - Use worker pools for request handling
  - Add circuit breakers for stability
