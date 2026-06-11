package com.lssgoo.feature.auth.domain.model

import kotlinx.serialization.Serializable

/**
 * Returned by `/auth/discover`. Lets the user pick a workspace before entering
 * a password — same UX as the web. Distinct from [TenantSummary] which represents
 * an *already-active* tenant.
 */
@Serializable
data class AuthTenant(
    val id: String,
    val name: String,
    val slug: String? = null,
    val avatarUrl: String? = null,
)
