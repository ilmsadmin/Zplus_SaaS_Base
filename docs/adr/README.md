# Architecture Decision Records (ADR)

## Overview

Architecture Decision Records (ADRs) document significant architectural decisions made during the development of Zplus SaaS Base. Each ADR captures the context, decision, and consequences of architectural choices.

## Table of Contents

- [ADR Template](#adr-template)
- [Decision Log](#decision-log)
- [Active ADRs](#active-adrs)

## ADR Template

```markdown
# ADR-XXX: [Decision Title]

## Status
[Proposed | Accepted | Rejected | Deprecated | Superseded]

## Context
What is the issue that we're seeing that is motivating this decision or change?

## Decision
What is the change that we're proposing and/or doing?

## Consequences
What becomes easier or more difficult to do because of this change?

## Alternatives Considered
What other options were considered and why were they rejected?

## References
- Links to relevant documents, discussions, or resources
```

## Decision Log

| ADR | Date | Title | Status |
|-----|------|-------|--------|
| [001](./001-multi-tenant-architecture.md) | 2025-01-15 | Multi-Tenant Architecture Strategy | Accepted |
| [002](./002-microservices-vs-monolith.md) | 2025-01-20 | Microservices vs Monolith | Accepted |
| [003](./003-database-per-tenant.md) | 2025-01-25 | Database Per Tenant Strategy | Accepted |
| [004](./004-graphql-federation.md) | 2025-02-01 | GraphQL Federation | Accepted |
| [005](./005-authentication-strategy.md) | 2025-02-05 | Authentication Strategy | Accepted |
| [006](./006-frontend-framework.md) | 2025-02-10 | Frontend Framework Choice | Accepted |
| [007](./007-container-orchestration.md) | 2025-02-15 | Container Orchestration Platform | Accepted |
| [008](./008-monitoring-stack.md) | 2025-02-20 | Monitoring and Observability Stack | Accepted |
| [009](./009-custom-domain-support.md) | 2025-08-07 | Custom Domain Support | Accepted |

## Active ADRs

### [ADR-001: Multi-Tenant Architecture Strategy](./001-multi-tenant-architecture.md)
**Status**: Accepted  
**Decision**: Implement schema-per-tenant for PostgreSQL and database-per-tenant for MongoDB  
**Impact**: High - Fundamental to the entire system architecture

### [ADR-002: Microservices vs Monolith](./002-microservices-vs-monolith.md)
**Status**: Accepted  
**Decision**: Start with modular monolith, evolve to microservices  
**Impact**: High - Affects development workflow and deployment strategy

### [ADR-003: Database Per Tenant Strategy](./003-database-per-tenant.md)
**Status**: Accepted  
**Decision**: Mixed approach - PostgreSQL schemas, MongoDB databases  
**Impact**: High - Critical for data isolation and compliance

### [ADR-004: GraphQL Federation](./004-graphql-federation.md)
**Status**: Accepted  
**Decision**: Use GraphQL Federation for API composition  
**Impact**: Medium - Affects API design and client integration

### [ADR-005: Authentication Strategy](./005-authentication-strategy.md)
**Status**: Accepted  
**Decision**: Keycloak for identity management with Casbin for authorization  
**Impact**: High - Affects security architecture and user experience

### [ADR-006: Frontend Framework Choice](./006-frontend-framework.md)
**Status**: Accepted  
**Decision**: Next.js 14 with App Router and Apollo Client  
**Impact**: Medium - Affects development speed and user experience

### [ADR-007: Container Orchestration Platform](./007-container-orchestration.md)
**Status**: Accepted  
**Decision**: Kubernetes on AWS EKS  
**Impact**: High - Affects scalability, deployment, and operations

### [ADR-008: Monitoring Stack](./008-monitoring-stack.md)
**Status**: Accepted  
**Decision**: Prometheus + Grafana + Loki for observability  
**Impact**: Medium - Affects operational visibility and debugging

## Review Process

ADRs should be reviewed and updated regularly:

1. **Quarterly Review**: Every 3 months, review all active ADRs
2. **Project Milestones**: Review relevant ADRs at major milestones
3. **Technology Changes**: Update ADRs when underlying technologies change
4. **Performance Issues**: Review related ADRs when performance problems arise

## Creating New ADRs

1. Copy the template above
2. Fill in all sections thoroughly
3. Discuss with the team
4. Get approval from technical leads
5. Update the decision log
6. Commit to repository

## Guidelines

- **Be Specific**: Include concrete details and reasoning
- **Consider Alternatives**: Document what was considered and why rejected
- **Think Long-term**: Consider maintenance and evolution
- **Include Metrics**: When possible, include performance or business metrics
- **Reference Sources**: Link to research, benchmarks, or discussions

### [ADR-009: Custom Domain Support](./009-custom-domain-support.md)
**Status**: Accepted  
**Decision**: Implement custom domain support với DNS verification và automatic SSL  
**Impact**: High - Enables white-label capabilities for enterprise customers

---

**Last Updated**: August 7, 2025  
**Next Review**: November 7, 2025
