# Spec: pkg-despecialization

## Purpose

Defines requirements for the `pkg-despecialization` capability.

## Requirements

### Requirement: No business-specific scope or identity methods on Actor

`pkg/actor` SHALL NOT define any business-specific scope key constants, convenience methods (`TenantID()`, `OrganizationID()`, etc.), or request-scope dimension bags (`Scope(key)` / `SetScope()`). The `Attrs() map[string]string` serves as the open extension mechanism.

#### Scenario: No scope key constants in pkg/actor

- **WHEN** `pkg/actor/user.go` is inspected
- **THEN** it SHALL NOT contain any exported `ScopeKey*` constants

#### Scenario: UserActor has no TenantID method

- **WHEN** code attempts to call `ua.TenantID()`
- **THEN** compilation SHALL fail (method removed)

#### Scenario: No Scope(key) method on Actor interface

- **WHEN** `Actor` interface is inspected
- **THEN** it SHALL NOT expose `Scope(key string) string`

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
