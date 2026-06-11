package com.lssgoo.feature.auth.data.dto

import com.lssgoo.feature.auth.domain.model.Session
import kotlinx.serialization.Serializable

/**
 * Wire types for the auth endpoints. We keep them separate from domain models
 * so backend rename/add-field doesn't ripple into use cases.
 *
 * The Go backend at `internal/modules/auth/handler.go` defines the canonical
 * shapes; keep this file in sync.
 */

@Serializable
data class DiscoverRequest(val email: String)

@Serializable
data class DiscoverResponse(
    val tenants: List<DiscoverTenantDto> = emptyList(),
)

@Serializable
data class DiscoverTenantDto(
    val id: String,
    val name: String,
    val slug: String? = null,
    val avatarUrl: String? = null,
)

@Serializable
data class LoginRequest(
    val email: String,
    val password: String,
    val tenantId: String? = null,
)

@Serializable
data class LoginResponse(
    val accessToken: String,
    val refreshToken: String,
    val accessTokenExpiresAt: String? = null,
    val refreshTokenExpiresAt: String? = null,
    val session: Session,
)
