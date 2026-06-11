package com.lssgoo.core.network

import com.lssgoo.core.storage.TokenStorage
import io.github.aakira.napier.Napier
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsText
import io.ktor.http.HttpStatusCode
import kotlinx.serialization.SerializationException
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

/**
 * Exchanges a refresh token for a new access/refresh pair.
 *
 * Kept separate from [ApiClient] because:
 *  - It must NOT recurse through ApiClient's 401-retry loop.
 *  - It needs its own minimal serialization to stay decoupled from auth-feature DTOs.
 *
 * The endpoint mirrors `internal/modules/auth/handler.go::POST /auth/refresh`.
 */
class TokenRefresher(
    private val httpClient: HttpClient,
    private val tokenStorage: TokenStorage,
) {
    private val json: Json = HttpClientFactory.json

    @Serializable
    private data class RefreshRequest(val refreshToken: String)

    @Serializable
    private data class RefreshPayload(
        val accessToken: String,
        val refreshToken: String,
        val accessTokenExpiresAt: String? = null,
        val refreshTokenExpiresAt: String? = null,
    )

    /** Returns true if tokens were refreshed and stored; false otherwise (caller should treat as logged out). */
    suspend fun refreshIfPossible(): Boolean {
        val refreshToken = tokenStorage.refreshToken() ?: return false
        return try {
            val response = httpClient.post("/auth/refresh") {
                setBody(RefreshRequest(refreshToken))
            }
            if (response.status != HttpStatusCode.OK) {
                Napier.w(tag = "TokenRefresher", message = "refresh returned ${response.status}")
                tokenStorage.clear()
                return false
            }

            val raw = response.bodyAsText()
            val envelope = json.decodeFromString(
                ApiResponse.serializer(RefreshPayload.serializer()),
                raw,
            )
            val payload = envelope.data
            if (!envelope.success || payload == null) {
                tokenStorage.clear()
                return false
            }
            tokenStorage.save(
                accessToken = payload.accessToken,
                refreshToken = payload.refreshToken,
                accessTokenExpiresAt = payload.accessTokenExpiresAt,
                refreshTokenExpiresAt = payload.refreshTokenExpiresAt,
            )
            true
        } catch (e: SerializationException) {
            Napier.e(tag = "TokenRefresher", message = "decode failed", throwable = e)
            false
        } catch (t: Throwable) {
            Napier.e(tag = "TokenRefresher", message = "refresh failed", throwable = t)
            false
        }
    }
}
