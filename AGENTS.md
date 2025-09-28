# ADMUX Project Handover Document

## Project Overview

**ADMUX** is an advertising ADX (Ad Exchange) bidding platform built using the OpenRTB standard protocol. The platform serves as a real-time bidding system that connects SSPs (Supply-Side Platforms) with DSPs (Demand-Side Platforms) through standardized OpenRTB protocol communication.

### Key Components

1. **ADX Server** (`cmd/adx_server/`)
   - Main bidding engine server
   - Runs on port 8080
   - Handles bid requests and orchestrates the bidding process

2. **Tracking Server** (`cmd/trcking_server/`)
   - Event tracking and analytics server
   - Runs on port 8081
   - Handles impression, click, and conversion tracking

3. **Configuration Management** (`internal/config/`)
   - Server configuration management
   - Default configuration with timeout settings
   - Extensible for production deployment

4. **Utility Functions** (`pkg/utils/`)
   - Random string generation for IDs and tokens
   - Cryptographic utilities

## Architecture

### Core Flow
1. **SSP Adapter Layer** (`internal/ssp/` - Planned)
   - Converts SSP-specific bid data to standardized OpenRTB BidReqCtx
   - Supports major SSPs: Xiaomi, Kuaishou, ByteDance, Maoyan, etc.

2. **ADX Core Engine** (`internal/adxcore/` - Planned)
   - **BidReqCtx**: Bidding request context
     - SSP information
     - BidRequest with implementations
     - ServerCtx with Redis caching, audience targeting, cookie mapping, bidder management
   - **Candidate**: BidResponse container

3. **DSP Adapter Layer** (`internal/bidder/` - Planned)
   - **BidderCtx**: Bidder context
     - BidderInfo (address, QPS limits, etc.)
     - BidRequest/BidResponse handling

4. **Response Packager** (Planned)
   - Converts OpenRTB responses to SSP-specific formats

### Technical Stack
- **Language**: Go 1.24.5
    - use 'any' replace 'interface{}'
- **Protocol**: OpenRTB standard
- **Dependencies**: Google Protobuf v1.36.9
- **Build System**: Makefile
- **Http Gin framework**

## Development Setup

### Prerequisites
- Go 1.24.5+
- Protocol Buffer compiler (protoc)
- Go protobuf plugin

### Build Commands
```bash
# Build both services
make build

# Build ADX server only
make build-adx

# Build Tracking server only
make build-tracking

# Generate protobuf code
make proto

# Run tests
make test

# Clean build artifacts
make clean
```

### Running Services
```bash
# Start ADX server
make run-adx

# Start Tracking server
make run-tracking
```

## Project Structure
```
├── cmd/
│   ├── adx_server/         # Main ADX bidding server
│   └── trcking_server/     # Tracking and analytics server
├── internal/
│   ├── config/             # Configuration management
│   ├── ssp/                # SSP adapter layer (planned)
│   ├── adx/                # Main ADX logic (planned)
│   ├── adxcore/            # Core bidding context (planned)
│   └── bidder/             # DSP bidder adapters (planned)
├── pkg/
│   └── utils/              # Utility functions
├── api/
│   └── idl/
│       └── openrtb/        # OpenRTB protocol definitions
└── bin/                    # Built binaries
```

## Key Features (Planned/In Development)

### 1. SSP Integration
- Multi-SSP support with adapter pattern
- Protocol conversion from SSP-specific to OpenRTB
- Rate limiting and QPS management

### 2. Bidding Engine
- Real-time bidding orchestration
- Bidder recall and selection
- Concurrent bidder request handling
- Protocol transformation (OpenRTB ↔ Bidder-specific)

### 3. Audience Targeting
- User segmentation and audience packages
- Cookie mapping and user identification
- Redis-based caching for performance

### 4. Response Processing
- Candidate ranking (coarse/fine/mixed ranking)
- Response packaging for SSP requirements
- Win notification handling

## Development Guidelines

### Code Organization
- Follow Go standard project layout
- Use interfaces for SSP and DSP adapters
- Implement proper error handling and logging
- Maintain OpenRTB protocol compliance

### Testing Strategy
- Unit tests for individual components
- Integration tests for bidding flow
- Performance testing for high QPS scenarios
- Protocol compliance validation

### Deployment Considerations
- Containerized deployment recommended
- Horizontal scaling for high traffic
- Redis cluster for distributed caching
- Monitoring and metrics collection

## Future Development Areas

1. **SSP Adapters Implementation**
   - Complete Xiaomi, Kuaishou, ByteDance adapters
   - Standardized adapter interface

2. **Bidder Integration**
   - Multiple DSP bidder support
   - Bidder health monitoring
   - Dynamic bidder configuration

3. **Advanced Features**
   - 预留pctr预估服务能力 

4. **Operational Excellence**
   - Comprehensive monitoring
   - Alerting system
   - Performance optimization
   - Disaster recovery procedures

5. **设计如果相关包中包含README相关的内容，阅读并理解其内容** 确保实现符合其设计约束
    - 每次改动完成后，一定记得完成相关README/Makefile 等内容的更新

## Critical Dependencies

- **Google Protobuf**: Protocol serialization
- **Redis**: Caching and session management
- **OpenRTB Protocol**: Industry standard compliance

## Known Issues & Technical Debt

- Current implementation is minimal (placeholder servers)
- SSP and DSP adapters need to be implemented
- Configuration system needs enhancement for production
- Missing comprehensive test suite
- No monitoring or logging infrastructure

## Contact & Support

- Project documentation: README.md
- Build system: Makefile
- Protocol definitions: api/idl/openrtb/

---
*This document serves as a comprehensive handover guide for new developers joining the ADMUX project. Last updated: 2025-09-21*

