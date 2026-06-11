package com.lssgoo.core.network

import com.lssgoo.core.result.AppError
import com.lssgoo.core.result.AppResult
import com.lssgoo.core.result.ErrorCode
import com.lssgoo.core.storage.TokenStorage
import io.github.aakira.napier.Napier
import io.ktor.client.HttpClient
import io.ktor.client.request.HttpRequestBuilder
import io.ktor.client.request.delete
import io.ktor.client.request.get
import io.ktor.client.request.header
import io.ktor.client.request.patch
import io.ktor.client.request.post
import io.ktor.client.request.put
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse
import io.ktor.client.statement.bodyAsText
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.isSuccess
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.serialization.KSerializer
import kotlinx.serialization.SerializationException
import kotlinx.serialization.json.Json

/**
 * Wraps [HttpClient] to add:
 *  - Bearer-token injection from [TokenStorage]
 *  - One-shot refresh on 401 (mirrors the Next.js proxy at `src/app/api/v1/[[...path]]/route.ts`)
 *  - Response-envelope unwrapping into [AppResult]
 *
 * The verb helpers take an explicit [KSerializer]. We deliberately avoid
 * `inline reified` because mixing inline + private state + suspend coroutines
 * trips visibility errors across KMP source sets. Passing the serializer keeps
 * the type machinery honest and the call sites only one line longer:
 *
 *     api.get(LoginResponse.serializer(), "/auth/me")
 */
class ApiClient(
    private val client: HttpClient,
    private val tokenStorage: TokenStorage,
    private val refresher: TokenRefresher,
) {
    private val refreshMutex = Mutex()
    private val json: Json = HttpClientFactory.json

    suspend fun <T> get(
        serializer: KSerializer<T>,
        path: String,
        block: HttpRequestBuilder.() -> Unit = {},
    ): AppResult<T> = execute(serializer, path) {
        client.get(path) { applyAuth(this); block(this) }
    }

    suspend fun <T> post(
        serializer: KSerializer<T>,
        path: String,
        body: Any? = null,
        block: HttpRequestBuilder.() -> Unit = {},
    ): AppResult<T> = execute(serializer, path) {
        client.post(path) {
            applyAuth(this)
            if (body != null) setBody(body)
            block(this)
        }
    }

    suspend fun <T> put(
        serializer: KSerializer<T>,
        path: String,
        body: Any? = null,
        block: HttpRequestBuilder.() -> Unit = {},
    ): AppResult<T> = execute(serializer, path) {
        client.put(path) {
            applyAuth(this)
            if (body != null) setBody(body)
            block(this)
        }
    }

    suspend fun <T> patch(
        serializer: KSerializer<T>,
        path: String,
        body: Any? = null,
        block: HttpRequestBuilder.() -> Unit = {},
    ): AppResult<T> = execute(serializer, path) {
        client.patch(path) {
            applyAuth(this)
            if (body != null) setBody(body)
            block(this)
        }
    }

    suspend fun <T> delete(
        serializer: KSerializer<T>,
        path: String,
        block: HttpRequestBuilder.() -> Unit = {},
    ): AppResult<T> = execute(serializer, path) {
        client.delete(path) { applyAuth(this); block(this) }
    }

    /**
     * Core request loop:
     *  1. Execute the call.
     *  2. If 401 and the path is eligible, refresh once and retry.
     *  3. Decode the envelope into [AppResult].
     */
    private suspend fun <T> execute(
        serializer: KSerializer<T>,
        path: String,
        call: suspend () -> HttpResponse,
    ): AppResult<T> {
        return try {
            var response = call()
            if (response.status == HttpStatusCode.Unauthorized && shouldAttemptRefresh(path)) {
                val refreshed = refreshMutex.withLock { refresher.refreshIfPossible() }
                if (refreshed) {
                    response = call()
                }
            }
            decode(serializer, response)
        } catch (t: Throwable) {
            Napier.e(tag = "ApiClient", message = "$path failed", throwable = t)
            AppResult.Failure(
                AppError(
                    code = if (t is SerializationException) ErrorCode.INTERNAL else ErrorCode.NETWORK,
                    message = t.message ?: "Network error",
                    cause = t,
                ),
            )
        }
    }

    private fun shouldAttemptRefresh(path: String): Boolean {
        // Auth endpoints can't recover by refreshing; would loop.
        return !path.startsWith("/auth/") || path == "/auth/me"
    }

    private suspend fun applyAuth(builder: HttpRequestBuilder) {
        val token = tokenStorage.accessToken()
        if (!token.isNullOrBlank()) {
            builder.header(HttpHeaders.Authorization, "Bearer $token")
        }
    }

    private suspend fun <T> decode(
        serializer: KSerializer<T>,
        response: HttpResponse,
    ): AppResult<T> {
        val raw = response.bodyAsText()
        if (raw.isBlank()) {
            return if (response.status.isSuccess()) {
                @Suppress("UNCHECKED_CAST")
                AppResult.Success(Unit as T)
            } else {
                AppResult.Failure(httpStatusToError(response.status))
            }
        }

        val envelopeSerializer = ApiResponse.serializer(serializer)
        val envelope: ApiResponse<T> = try {
            json.decodeFromString(envelopeSerializer, raw)
        } catch (e: SerializationException) {
            return AppResult.Failure(
                AppError(
                    code = ErrorCode.INTERNAL,
                    message = "Malformed response: ${e.message}",
                    cause = e,
                ),
            )
        }

        return when {
            envelope.success && envelope.data != null -> AppResult.Success(envelope.data)
            envelope.success && envelope.data == null -> {
                @Suppress("UNCHECKED_CAST")
                AppResult.Success(Unit as T)
            }
            else -> AppResult.Failure(
                AppError(
                    code = ErrorCode.fromRaw(envelope.error?.code),
                    message = envelope.error?.message
                        ?: envelope.message
                        ?: "Request failed (${response.status.value})",
                    details = envelope.error?.details.orEmpty(),
                ),
            )
        }
    }

    private fun httpStatusToError(status: HttpStatusCode): AppError = AppError(
        code = when (status) {
            HttpStatusCode.Unauthorized -> ErrorCode.UNAUTHORIZED
            HttpStatusCode.Forbidden -> ErrorCode.FORBIDDEN
            HttpStatusCode.NotFound -> ErrorCode.NOT_FOUND
            HttpStatusCode.Conflict -> ErrorCode.CONFLICT
            HttpStatusCode.TooManyRequests -> ErrorCode.RATE_LIMITED
            HttpStatusCode.RequestTimeout -> ErrorCode.TIMEOUT
            else -> if (status.value >= 500) ErrorCode.INTERNAL else ErrorCode.UNKNOWN
        },
        message = status.description,
    )
}
