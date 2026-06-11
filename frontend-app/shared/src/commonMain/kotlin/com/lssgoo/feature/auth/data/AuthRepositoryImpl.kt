package com.lssgoo.feature.auth.data

import com.lssgoo.core.result.AppResult
import com.lssgoo.core.storage.SessionCache
import com.lssgoo.core.storage.TokenStorage
import com.lssgoo.feature.auth.domain.AuthRepository
import com.lssgoo.feature.auth.domain.model.AuthTenant
import com.lssgoo.feature.auth.domain.model.Session
import kotlinx.coroutines.flow.StateFlow

/**
 * The only place that touches both [AuthApi] and storage. Responsibilities:
 *  - Decide what's "successful" enough to persist.
 *  - Keep [TokenStorage] and [SessionCache] in lockstep.
 *  - Convert DTOs to domain models.
 *
 * Keeping persistence here (vs in a use case) means accidental "I logged in but
 * forgot to save tokens" bugs become impossible: every login path goes through
 * this class.
 */
class AuthRepositoryImpl(
    private val api: AuthApi,
    private val tokenStorage: TokenStorage,
    private val sessionCache: SessionCache,
) : AuthRepository {

    override val session: StateFlow<Session?> = sessionCache.session

    override suspend fun discoverTenants(email: String): AppResult<List<AuthTenant>> =
        api.discover(email).map { res ->
            res.tenants.map { dto ->
                AuthTenant(
                    id = dto.id,
                    name = dto.name,
                    slug = dto.slug,
                    avatarUrl = dto.avatarUrl,
                )
            }
        }

    override suspend fun login(
        email: String,
        password: String,
        tenantId: String?,
    ): AppResult<Session> {
        val result = api.login(email, password, tenantId)
        result.onSuccess { payload ->
            tokenStorage.save(
                accessToken = payload.accessToken,
                refreshToken = payload.refreshToken,
                accessTokenExpiresAt = payload.accessTokenExpiresAt,
                refreshTokenExpiresAt = payload.refreshTokenExpiresAt,
            )
            sessionCache.update(payload.session)
        }
        return result.map { it.session }
    }

    override suspend fun fetchSession(): AppResult<Session> {
        val result = api.me()
        result.onSuccess { sessionCache.update(it) }
        return result
    }

    override suspend fun logout(): AppResult<Unit> {
        // Even if the server call fails (network down, token already expired)
        // we still wipe local state — the user clicked Logout, that's enough.
        val result = api.logout()
        tokenStorage.clear()
        sessionCache.clear()
        return result
    }

    override suspend fun hasCredentials(): Boolean = tokenStorage.hasTokens()
}
