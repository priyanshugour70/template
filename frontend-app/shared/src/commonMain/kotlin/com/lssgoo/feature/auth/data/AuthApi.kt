package com.lssgoo.feature.auth.data

import com.lssgoo.core.network.ApiClient
import com.lssgoo.core.result.AppResult
import com.lssgoo.feature.auth.data.dto.DiscoverRequest
import com.lssgoo.feature.auth.data.dto.DiscoverResponse
import com.lssgoo.feature.auth.data.dto.LoginRequest
import com.lssgoo.feature.auth.data.dto.LoginResponse
import com.lssgoo.feature.auth.domain.model.Session
import kotlinx.serialization.builtins.serializer

/**
 * Thin wrapper around [ApiClient]. One responsibility per method, no business
 * logic. Repositories orchestrate; this just talks HTTP.
 *
 * Endpoint paths mirror the Go backend's `internal/modules/auth/handler.go`.
 */
class AuthApi(private val client: ApiClient) {

    suspend fun discover(email: String): AppResult<DiscoverResponse> =
        client.post(
            serializer = DiscoverResponse.serializer(),
            path = "/auth/discover",
            body = DiscoverRequest(email),
        )

    suspend fun login(
        email: String,
        password: String,
        tenantId: String?,
    ): AppResult<LoginResponse> = client.post(
        serializer = LoginResponse.serializer(),
        path = "/auth/login",
        body = LoginRequest(email = email, password = password, tenantId = tenantId),
    )

    suspend fun me(): AppResult<Session> =
        client.get(serializer = Session.serializer(), path = "/auth/me")

    suspend fun logout(): AppResult<Unit> =
        client.post(serializer = Unit.serializer(), path = "/auth/logout")
}
