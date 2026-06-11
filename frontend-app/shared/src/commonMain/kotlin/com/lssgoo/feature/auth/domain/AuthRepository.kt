package com.lssgoo.feature.auth.domain

import com.lssgoo.core.result.AppResult
import com.lssgoo.feature.auth.domain.model.AuthTenant
import com.lssgoo.feature.auth.domain.model.Session
import kotlinx.coroutines.flow.StateFlow

/**
 * Domain interface for everything auth-related. Presentation depends on this,
 * never on the concrete [com.lssgoo.feature.auth.data.AuthRepositoryImpl] —
 * making it trivial to swap for a fake in tests.
 */
interface AuthRepository {
    /** Live session as last known by [com.lssgoo.core.storage.SessionCache]. */
    val session: StateFlow<Session?>

    /** Returns tenants the email is a member of so the user can pick one. */
    suspend fun discoverTenants(email: String): AppResult<List<AuthTenant>>

    /** Authenticate. On success, tokens are persisted and [session] is updated. */
    suspend fun login(email: String, password: String, tenantId: String?): AppResult<Session>

    /** Fetch the current session from the server. Used at app start. */
    suspend fun fetchSession(): AppResult<Session>

    /** Revoke tokens server-side + clear local storage. */
    suspend fun logout(): AppResult<Unit>

    /** True if any access token is on disk — does NOT prove validity. */
    suspend fun hasCredentials(): Boolean
}
