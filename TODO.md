# Bourbon Framework TODO

**Vision**: Build a Go web framework with feature parity to Django, Laravel, and Spring Boot.

**Current Version**: 0.0.1  
**Target Version**: 1.0.0 (Production Ready)

---

## ✅ Current Features (v0.0.1)

- [x] Django-inspired app structure
- [x] GORM integration (PostgreSQL, MySQL, SQLite)
- [x] Smart migrations with auto-detection
- [x] RESTful routing with groups and parameters
- [x] Middleware system (Logger, Recovery, CORS)
- [x] Template engine with auto-reload
- [x] Static file serving
- [x] Structured logging (Uber Zap)
- [x] CLI scaffolding (bourbon new, create:app)
- [x] Error storage in database
- [x] Async job dispatcher interface
- [x] Named middleware registry
- [x] Path normalization in routes

---

## Phase 1: Core Stability & Essential Features (v0.1.0)

**Goal**: Production-ready foundation with critical web application features.  
**Priority**: Critical

### Authentication & Authorization
- [ ] **Authentication System**
  - [ ] User model with password hashing (bcrypt)
  - [ ] Login/Logout handlers
  - [ ] Session management (cookie-based & token-based)
  - [ ] Remember me functionality
  - [ ] Password reset via email
  - [ ] Email verification
  - [ ] Multi-factor authentication (TOTP)
  
- [ ] **Authorization System**
  - [ ] Permission-based access control
  - [ ] Role-based access control (RBAC)
  - [ ] Policy-based authorization (like Laravel Gates)
  - [ ] Middleware: `RequireAuth`, `RequireRole`, `RequirePermission`
  - [ ] Context helpers: `c.User()`, `c.Can("permission")`

### Request Validation
- [ ] **Validation Framework**
  - [ ] Struct tag-based validation (inspired by go-playground/validator)
  - [ ] Custom validation rules
  - [ ] Validation error messages with i18n support
  - [ ] Form validation
  - [ ] JSON validation
  - [ ] Query parameter validation
  - [ ] File upload validation
  - [ ] `c.Validate()` method in Context

### Session Management
- [ ] **Session System**
  - [ ] Session store interface
  - [ ] Cookie-based session store
  - [ ] Database session store
  - [ ] Redis session store
  - [ ] Flash messages
  - [ ] CSRF protection
  - [ ] Session middleware

### Security Enhancements
- [ ] **Security Features**
  - [ ] CSRF protection middleware
  - [ ] XSS protection (automatic escaping in templates)
  - [ ] SQL injection prevention (GORM already helps)
  - [ ] Rate limiting middleware
  - [ ] IP whitelisting/blacklisting
  - [ ] Security headers middleware (HSTS, X-Frame-Options, etc.)
  - [ ] Content Security Policy (CSP)
  - [ ] Secure cookie handling

### Testing Framework
- [ ] **Testing Support**
  - [ ] Test database creation/teardown
  - [ ] HTTP test helpers
  - [ ] Mock context creation
  - [ ] Factory pattern for models
  - [ ] Database seeding for tests
  - [ ] Integration test utilities
  - [ ] Example test suite

---

## Phase 2: Developer Experience & Productivity (v0.2.0)

**Goal**: Tools and features that accelerate development.  
**Priority**: High

### Database Features
- [ ] **Advanced ORM Features**
  - [ ] Database seeding system
  - [ ] Factory pattern for generating test data
  - [ ] Model observers/hooks (BeforeCreate, AfterUpdate, etc.)
  - [ ] Soft delete support (already in BaseModel, expand)
  - [ ] Database transactions helper
  - [ ] Query scopes (reusable query logic)
  - [ ] Polymorphic relationships
  - [ ] Database connection pooling tuning

- [ ] **Migration Enhancements**
  - [ ] Migration squashing
  - [ ] Data migrations (not just schema)
  - [ ] Migration testing (dry-run mode)
  - [ ] Cross-database migration compatibility checker

### CLI Enhancements
- [ ] **Enhanced CLI Commands**
  - [ ] `bourbon make:controller` - Generate controller
  - [ ] `bourbon make:model` - Generate model
  - [ ] `bourbon make:middleware` - Generate middleware
  - [ ] `bourbon make:migration` - Enhanced with templates
  - [ ] `bourbon make:seeder` - Generate seeder
  - [ ] `bourbon make:factory` - Generate factory
  - [ ] `bourbon make:test` - Generate test file
  - [ ] `bourbon make:api` - Generate API resource
  - [ ] `bourbon db:seed` - Run seeders
  - [ ] `bourbon db:fresh` - Drop and recreate database
  - [ ] `bourbon routes:list` - Show all registered routes
  - [ ] `bourbon cache:clear` - Clear application cache

### Form Handling
- [ ] **Form Builder**
  - [ ] HTML form generator from structs
  - [ ] Form field rendering helpers
  - [ ] CSRF token automatic injection
  - [ ] Old input repopulation on errors
  - [ ] File upload handling
  - [ ] Multi-part form support

### API Development
- [ ] **API Resources**
  - [ ] JSON resource transformers (like Laravel Resources)
  - [ ] Pagination helpers
  - [ ] API versioning support
  - [ ] Rate limiting for APIs
  - [ ] API key authentication
  - [ ] OpenAPI/Swagger documentation generation
  - [ ] HATEOAS links support

### Localization
- [ ] **Internationalization (i18n)**
  - [ ] Translation file loading (JSON/YAML)
  - [ ] `__()` translation function
  - [ ] Pluralization support
  - [ ] Locale detection from request
  - [ ] Date/time formatting per locale
  - [ ] Number formatting per locale
  - [ ] Template translation functions

---

## Phase 3: Advanced Features & Scalability (v0.3.0)

**Goal**: Enterprise-grade features for large-scale applications.  
**Priority**: High

### Caching System
- [ ] **Cache Framework**
  - [ ] Cache interface (Get, Set, Delete, Flush)
  - [ ] Memory cache driver
  - [ ] Redis cache driver
  - [ ] File cache driver
  - [ ] Database cache driver
  - [ ] Cache tags
  - [ ] Cache middleware for routes
  - [ ] Template fragment caching
  - [ ] Query result caching
  - [ ] `c.Cache()` helper in Context

### Queue & Job System
- [ ] **Enhanced Queue System**
  - [ ] Database-backed queue driver
  - [ ] Redis queue driver
  - [ ] RabbitMQ driver
  - [ ] Job worker process
  - [ ] Job retry logic with exponential backoff
  - [ ] Job prioritization
  - [ ] Failed job handling
  - [ ] Dead letter queue
  - [ ] Job monitoring dashboard
  - [ ] Scheduled jobs (cron-like)
  - [ ] Job middleware
  - [ ] Job batching

### Event System
- [ ] **Event & Listeners**
  - [ ] Event dispatcher
  - [ ] Event listeners registration
  - [ ] Synchronous events
  - [ ] Asynchronous events (queued listeners)
  - [ ] Event discovery
  - [ ] Built-in events (UserCreated, ModelSaved, etc.)

### WebSocket Support
- [ ] **Real-time Features**
  - [ ] WebSocket server integration
  - [ ] Broadcasting system
  - [ ] Channels (public, private, presence)
  - [ ] Event broadcasting to channels
  - [ ] Redis broadcaster
  - [ ] WebSocket authentication
  - [ ] Echo-like JavaScript client

### File Storage
- [ ] **Storage Abstraction**
  - [ ] Storage interface (Put, Get, Delete, Exists)
  - [ ] Local filesystem driver
  - [ ] S3-compatible driver (AWS S3, MinIO)
  - [ ] Google Cloud Storage driver
  - [ ] Azure Blob Storage driver
  - [ ] File visibility (public/private)
  - [ ] Temporary URL generation
  - [ ] Stream support for large files
  - [ ] Image manipulation helpers

### Email System
- [ ] **Mail Framework**
  - [ ] Mail interface and drivers
  - [ ] SMTP driver
  - [ ] Mailgun driver
  - [ ] SendGrid driver
  - [ ] SES driver
  - [ ] Mail templates (HTML & text)
  - [ ] Markdown mail support
  - [ ] Mail queuing
  - [ ] Attachments support
  - [ ] Inline images

### Notification System
- [ ] **Notifications**
  - [ ] Notification channels (mail, database, SMS, Slack)
  - [ ] Notification interface
  - [ ] Database notification storage
  - [ ] Notification templates
  - [ ] Notification queuing
  - [ ] On-demand notifications

---

## Phase 4: Ecosystem & Tooling (v0.4.0)

**Goal**: Rich ecosystem with first-party packages and developer tools.  
**Priority**: Medium

### Admin Panel
- [ ] **Admin Dashboard** (Django Admin inspired)
  - [ ] Auto-generated CRUD interfaces from models
  - [ ] Customizable list views (filters, search, sorting)
  - [ ] Form generation for create/edit
  - [ ] Relationship handling in forms
  - [ ] Bulk actions
  - [ ] Export to CSV/Excel
  - [ ] Admin authentication
  - [ ] Admin permissions
  - [ ] Dashboard widgets
  - [ ] Action history/audit log

### Monitoring & Observability
- [ ] **Application Monitoring** (Spring Boot Actuator inspired)
  - [ ] Health check endpoints
  - [ ] Metrics collection (requests, response times, errors)
  - [ ] Prometheus metrics exporter
  - [ ] Database connection pool metrics
  - [ ] Memory usage metrics
  - [ ] CPU metrics
  - [ ] Custom metrics registration
  - [ ] `/metrics` endpoint
  - [ ] `/health` endpoint with custom checks

### Debugging & Development
- [ ] **Debug Tools**
  - [ ] Debug toolbar (SQL queries, timing, memory)
  - [ ] Query logging with explain
  - [ ] Request/response inspector
  - [ ] Hot reload for code changes
  - [ ] Interactive debugger
  - [ ] Performance profiling integration (pprof)

### Package System
- [ ] **Plugin Architecture**
  - [ ] Plugin interface
  - [ ] Plugin discovery
  - [ ] Plugin lifecycle hooks
  - [ ] Third-party package registry
  - [ ] Official packages repository

### GraphQL Support
- [ ] **GraphQL Integration**
  - [ ] GraphQL server (gqlgen integration)
  - [ ] Schema generation from models
  - [ ] Query resolvers
  - [ ] Mutation resolvers
  - [ ] Subscriptions (WebSocket)
  - [ ] DataLoader for N+1 prevention
  - [ ] GraphQL playground

---

## Phase 5: Enterprise & Production (v0.5.0 → v1.0.0)

**Goal**: Enterprise-ready with advanced deployment and scaling features.  
**Priority**: Medium

### Deployment & DevOps
- [ ] **Deployment Tools**
  - [ ] Docker support (Dockerfile generation)
  - [ ] Docker Compose templates
  - [ ] Kubernetes manifests generation
  - [ ] Health check for orchestrators
  - [ ] Graceful shutdown handling
  - [ ] Zero-downtime deployment support
  - [ ] Environment-based configuration
  - [ ] Secrets management integration (Vault, AWS Secrets Manager)

### Microservices Support
- [ ] **Distributed Systems**
  - [ ] Service discovery (Consul, etcd)
  - [ ] gRPC support
  - [ ] Circuit breaker pattern
  - [ ] Retry policies
  - [ ] Distributed tracing (OpenTelemetry, Jaeger)
  - [ ] Service mesh integration
  - [ ] API gateway patterns

### Advanced Database
- [ ] **Multi-tenancy Support**
  - [ ] Schema-based multi-tenancy
  - [ ] Database-based multi-tenancy
  - [ ] Tenant resolver middleware
  - [ ] Tenant-scoped queries

- [ ] **Read Replicas**
  - [ ] Master-slave configuration
  - [ ] Read/write splitting
  - [ ] Connection pool per replica

- [ ] **Sharding Support**
  - [ ] Horizontal sharding
  - [ ] Shard key configuration
  - [ ] Cross-shard queries

### Advanced Security
- [ ] **OAuth2 & OpenID Connect**
  - [ ] OAuth2 server implementation
  - [ ] OAuth2 client
  - [ ] Social authentication (Google, GitHub, Facebook)
  - [ ] JWT token handling
  - [ ] Refresh tokens
  - [ ] Scope-based authorization

- [ ] **API Security**
  - [ ] API key management
  - [ ] Rate limiting per user/key
  - [ ] IP-based access control
  - [ ] Webhook signature verification

### Performance Optimization
- [ ] **Performance Features**
  - [ ] Response compression (gzip, brotli)
  - [ ] HTTP/2 support
  - [ ] ETag generation
  - [ ] Last-Modified headers
  - [ ] Conditional requests
  - [ ] Asset pipeline (minification, bundling)
  - [ ] CDN integration helpers
  - [ ] Database query optimization analyzer

### Task Scheduling
- [ ] **Scheduler** (Laravel Task Scheduling inspired)
  - [ ] Cron-like task scheduling
  - [ ] Scheduled job registration
  - [ ] Task chaining
  - [ ] Task overlap prevention
  - [ ] Scheduled maintenance mode

---

## Phase 6: Documentation & Community (Ongoing)

**Goal**: Comprehensive documentation and thriving community.  
**Priority**: High

### Documentation
- [ ] **Complete Documentation**
  - [x] Getting started guide
  - [x] API reference
  - [ ] Tutorial series (blog, API, e-commerce)
  - [ ] Video tutorials
  - [ ] Architecture deep dive
  - [ ] Best practices guide
  - [ ] Security guide
  - [ ] Performance optimization guide
  - [ ] Deployment guide (AWS, GCP, Azure, DigitalOcean)
  - [ ] Upgrade guides
  - [ ] Contributing guide
  - [ ] FAQ

### Community
- [ ] **Community Building**
  - [ ] Official website
  - [ ] Community forum/Discord
  - [ ] Blog with updates
  - [ ] Newsletter
  - [ ] Showcase of apps built with Bourbon
  - [ ] Conference talks/presentations
  - [ ] Code of conduct
  - [ ] Contributor recognition

### Ecosystem Packages
- [ ] **Official Packages**
  - [ ] bourbon/auth - Authentication & authorization
  - [ ] bourbon/admin - Admin panel
  - [ ] bourbon/api - API development tools
  - [ ] bourbon/queue - Queue system
  - [ ] bourbon/cache - Caching framework
  - [ ] bourbon/mail - Email system
  - [ ] bourbon/storage - File storage
  - [ ] bourbon/websocket - WebSocket support
  - [ ] bourbon/graphql - GraphQL integration

---

## Comparison Matrix

### Feature Parity Status

| Feature | Django | Laravel | Spring Boot | Bourbon v0.0.1 | Target Phase |
|---------|:------:|:-------:|:-----------:|:--------------:|:------------:|
| **Core** |
| Routing | ✅ | ✅ | ✅ | ✅ | - |
| Middleware | ✅ | ✅ | ✅ | ✅ | - |
| Template Engine | ✅ | ✅ | ✅ | ✅ | - |
| ORM/Database | ✅ | ✅ | ✅ | ✅ | - |
| Migrations | ✅ | ✅ | ✅ | ✅ | - |
| CLI Tools | ✅ | ✅ | ✅ | ✅ | - |
| Static Files | ✅ | ✅ | ✅ | ✅ | - |
| **Authentication** |
| User Auth | ✅ | ✅ | ✅ | ❌ | Phase 1 |
| Sessions | ✅ | ✅ | ✅ | ❌ | Phase 1 |
| Permissions | ✅ | ✅ | ✅ | ❌ | Phase 1 |
| OAuth2 | ✅ | ✅ | ✅ | ❌ | Phase 5 |
| **Validation** |
| Form Validation | ✅ | ✅ | ✅ | ❌ | Phase 1 |
| Request Validation | ✅ | ✅ | ✅ | ❌ | Phase 1 |
| **Security** |
| CSRF Protection | ✅ | ✅ | ✅ | ❌ | Phase 1 |
| Rate Limiting | ✅ | ✅ | ✅ | ❌ | Phase 1 |
| XSS Protection | ✅ | ✅ | ✅ | ❌ | Phase 1 |
| **Testing** |
| Test Framework | ✅ | ✅ | ✅ | ⚠️ | Phase 1 |
| Factories | ✅ | ✅ | ✅ | ❌ | Phase 2 |
| Seeders | ✅ | ✅ | ✅ | ❌ | Phase 2 |
| **Admin** |
| Admin Panel | ✅ | ⚠️ | ❌ | ❌ | Phase 4 |
| **API** |
| REST Support | ✅ | ✅ | ✅ | ✅ | - |
| API Resources | ✅ | ✅ | ✅ | ❌ | Phase 2 |
| GraphQL | ⚠️ | ⚠️ | ⚠️ | ❌ | Phase 4 |
| OpenAPI/Swagger | ⚠️ | ✅ | ✅ | ❌ | Phase 2 |
| **Async** |
| Background Jobs | ✅ | ✅ | ✅ | ⚠️ | Phase 3 |
| Task Scheduling | ✅ | ✅ | ✅ | ❌ | Phase 5 |
| WebSockets | ⚠️ | ✅ | ✅ | ❌ | Phase 3 |
| **Caching** |
| Cache System | ✅ | ✅ | ✅ | ❌ | Phase 3 |
| **Storage** |
| File Storage | ✅ | ✅ | ✅ | ⚠️ | Phase 3 |
| Cloud Storage | ✅ | ✅ | ✅ | ❌ | Phase 3 |
| **Email** |
| Email System | ✅ | ✅ | ✅ | ❌ | Phase 3 |
| Email Templates | ✅ | ✅ | ✅ | ❌ | Phase 3 |
| **i18n** |
| Localization | ✅ | ✅ | ✅ | ❌ | Phase 2 |
| **Events** |
| Event System | ✅ | ✅ | ✅ | ❌ | Phase 3 |
| Notifications | ⚠️ | ✅ | ⚠️ | ❌ | Phase 3 |
| **Monitoring** |
| Health Checks | ⚠️ | ⚠️ | ✅ | ❌ | Phase 4 |
| Metrics | ⚠️ | ⚠️ | ✅ | ❌ | Phase 4 |
| **DI/IoC** |
| Dependency Injection | ⚠️ | ✅ | ✅ | ⚠️ | Phase 5 |

**Legend**: ✅ Full Support | ⚠️ Partial/Third-party | ❌ Not Available

---

## Contributing

We welcome contributions! If you want to work on any feature from this TODO list:

1. Check if an issue already exists
2. Create an issue describing what you want to work on
3. Fork the repository and create a branch
4. Submit a PR referencing the issue
5. See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines

---

## Priorities

### High Priority (Phase 1)
1. Authentication system
2. Validation framework
3. Session management
4. Security enhancements (CSRF, Rate limiting)
5. Testing framework improvements

### Medium Priority (Phase 2)
1. Database seeding & factories
2. Enhanced CLI commands
3. API resources & pagination
4. Form handling
5. Internationalization

### Standard Priority (Phases 3-4)
1. Caching framework
2. Enhanced queue system
3. Event system
4. WebSocket support
5. File storage abstraction
6. Admin panel
7. Monitoring & metrics

### Future Enhancements (Phase 5)
1. Microservices support
2. OAuth2 server
3. Multi-tenancy
4. Advanced database features

---

## Notes

- **Breaking Changes**: Expected in versions < 1.0.0
- **Semantic Versioning**: Will follow strictly from v1.0.0 onwards
- **Community Input**: Feature priorities may shift based on community feedback
- **Go Idioms**: All features will follow Go best practices and idioms
- **Performance**: Go's performance characteristics will be maintained
- **Type Safety**: Leverage Go's type system where applicable

---

**Maintainers**: [@ishubhamsingh2e](https://github.com/ishubhamsingh2e)

---

## Get Involved

- ⭐ Star the project on GitHub
- 🐛 Report bugs and request features
- 💬 Join our community discussions
- 📝 Contribute code or documentation
- 💰 Sponsor the project

**Together, we'll build the best Go web framework!** 🚀
