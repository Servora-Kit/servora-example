# Spec: pkg-despecialization

## Purpose

Defines requirements for the `pkg-despecialization` capability.

## Requirements

### Requirement: Actor scope keys are caller-defined, not framework-hardcoded

`pkg/actor` SHALL NOT define any business-specific scope key constants (e.g. `ScopeKeyTenantID`, `ScopeKeyOrganizationID`, `ScopeKeyProjectID`). The generic `Scope(key string) string` and `SetScope(key, value string)` API SHALL be the only scope access mechanism.

#### Scenario: No scope key constants in pkg/actor

- **WHEN** `pkg/actor/user.go` is inspected
- **THEN** it SHALL NOT contain any exported `ScopeKey*` constants

#### Scenario: Caller defines own scope keys

- **WHEN** a service needs tenant scope
- **THEN** the service SHALL define its own constant (e.g. `const ScopeKeyTenantID = "tenant_id"`) and use `actor.Scope("tenant_id")` / `actor.SetScope("tenant_id", id)`

### Requirement: No business-specific convenience methods on UserActor

`UserActor` SHALL NOT expose domain-specific convenience methods such as `TenantID()`, `SetTenantID()`, `OrganizationID()`, `SetOrganizationID()`, `ProjectID()`, `SetProjectID()`. These are syntactic sugar over generic `Scope()` / `SetScope()` and embed business assumptions.

#### Scenario: UserActor has no TenantID method

- **WHEN** code attempts to call `ua.TenantID()`
- **THEN** compilation SHALL fail (method removed)

### Requirement: Generic ScopeFromContext helper

`pkg/actor` SHALL provide a generic `ScopeFromContext(ctx context.Context, key string) (string, bool)` function that extracts a scope value from the actor in context by key.

#### Scenario: Scope value present

- **WHEN** an actor in context has scope key `"tenant_id"` set to `"abc-123"`
- **AND** `ScopeFromContext(ctx, "tenant_id")` is called
- **THEN** it SHALL return `("abc-123", true)`

#### Scenario: Scope value absent

- **WHEN** an actor in context has no scope key `"tenant_id"`
- **AND** `ScopeFromContext(ctx, "tenant_id")` is called
- **THEN** it SHALL return `("", false)`

#### Scenario: No actor in context

- **WHEN** no actor is in the context
- **AND** `ScopeFromContext(ctx, "tenant_id")` is called
- **THEN** it SHALL return `("", false)`

### Requirement: No legacy Metadata on UserActor

`UserActor` SHALL NOT have `Metadata map[string]string`, `Metadata()`, or `Meta()` fields/methods. `Attrs() map[string]string` serves the same purpose.

#### Scenario: Metadata field removed from UserActorParams

- **WHEN** `UserActorParams` is inspected
- **THEN** it SHALL NOT have a `Metadata` field

### Requirement: SystemActor ID is caller-provided

`SystemActor.ID()` SHALL return the ID as provided by the caller at construction time, without any automatic prefix (e.g. no `"system:"` prepended). The caller is responsible for providing the full ID string.

#### Scenario: SystemActor preserves caller ID

- **WHEN** `NewSystemActor("system:my-svc")` is called
- **THEN** `ID()` SHALL return `"system:my-svc"`

#### Scenario: SystemActor does not add prefix

- **WHEN** `NewSystemActor("my-svc")` is called
- **THEN** `ID()` SHALL return `"my-svc"` (not `"system:my-svc"`)

### Requirement: Authz middleware supports multi-actor-type principal construction

`pkg/authz` middleware SHALL dynamically construct the OpenFGA principal string based on `actor.Type()` and `actor.ID()`, using the pattern `string(a.Type()) + ":" + a.ID()`. It SHALL NOT hardcode `"user:"` prefix.

This logic SHALL reside in `pkg/authz/authz.go` (middleware layer), not in the `Authorizer` engine. The engine receives the fully constructed subject string.

#### Scenario: User actor principal

- **WHEN** a request from a user actor with Type `"user"` and ID `"alice"` is authorized
- **THEN** the middleware SHALL construct principal `"user:alice"` and pass it to `authorizer.IsAuthorized`

#### Scenario: Service actor principal

- **WHEN** a request from a service actor with Type `"service"` and ID `"order-svc"` is authorized
- **THEN** the middleware SHALL construct principal `"service:order-svc"` and pass it to `authorizer.IsAuthorized`

### Requirement: Authz middleware allows configurable non-checkable actor types

`pkg/authz` middleware SHALL NOT hardcode which actor types are rejected. By default, `anonymous` actors SHALL be rejected (no identity), but `user` and `service` actors SHALL both be allowed through to the `authorizer.IsAuthorized` call.

#### Scenario: Service actor passes authz check

- **WHEN** a service actor with ID `"order-svc"` makes a request to a CHECK operation
- **AND** the `Authorizer.IsAuthorized` returns `true`
- **THEN** the middleware SHALL allow the request

#### Scenario: Anonymous actor is rejected

- **WHEN** an anonymous actor makes a request to a CHECK operation
- **THEN** the middleware SHALL return 403 AUTHZ_DENIED without calling the `Authorizer`

### Requirement: Authz default object ID is configurable

`pkg/authz` SHALL use `"default"` as the fallback object ID when `IDField` is empty, but SHALL allow overriding this via `WithDefaultObjectID(id string)` option.

#### Scenario: Default fallback ID

- **WHEN** a rule has empty `IDField` and no `WithDefaultObjectID` is set
- **THEN** the object ID SHALL be `"default"`

#### Scenario: Custom fallback ID

- **WHEN** `WithDefaultObjectID("singleton")` is set
- **AND** a rule has empty `IDField`
- **THEN** the object ID SHALL be `"singleton"`
