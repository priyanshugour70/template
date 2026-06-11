package com.lssgoo.feature.auth.domain.model

import kotlinx.serialization.Serializable

/**
 * Top-level result of a successful login / refresh / "who am I" call. Mirrors the
 * backend's `SessionResponse` exactly so DTOs and domain models are the same shape.
 *
 * Persisted in [com.lssgoo.core.storage.SessionCache] so the UI shell (sidebar,
 * user menu) can render before any network call resolves, matching the web's
 * Zustand-hydration pattern.
 */
@Serializable
data class Session(
    val user: SessionUser,
    val tenant: TenantSummary,
    val activeOrganization: OrganizationSummary? = null,
    val organizations: List<OrganizationSummary> = emptyList(),
    val permissions: List<String> = emptyList(),
    val roles: List<String> = emptyList(),
)

@Serializable
data class SessionUser(
    val id: String,
    val email: String,
    val firstName: String? = null,
    val lastName: String? = null,
    val displayName: String? = null,
    val avatarUrl: String? = null,
    val isSuperAdmin: Boolean = false,
) {
    val fullName: String
        get() = displayName
            ?: listOfNotNull(firstName, lastName).joinToString(" ").ifBlank { email }
}

@Serializable
data class TenantSummary(
    val id: String,
    val name: String,
    val slug: String? = null,
)

@Serializable
data class OrganizationSummary(
    val id: String,
    val name: String,
    val slug: String? = null,
)
