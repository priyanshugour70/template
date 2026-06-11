package com.lssgoo.core.network

import com.lssgoo.core.config.AppConfig
import io.github.aakira.napier.Napier
import io.ktor.client.HttpClient
import io.ktor.client.HttpClientConfig
import io.ktor.client.engine.HttpClientEngine
import io.ktor.client.plugins.DefaultRequest
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.logging.LogLevel
import io.ktor.client.plugins.logging.Logger
import io.ktor.client.plugins.logging.Logging
import io.ktor.client.request.header
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.contentType
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json

/**
 * Single source of truth for the Ktor [HttpClient] used by every feature. Centralising
 * the JSON config, headers, timeouts and logging here means new features get correct
 * behavior for free.
 *
 * The platform engine ([HttpClientEngine]) is injected via Koin — OkHttp on Android,
 * Darwin on iOS — keeping this factory engine-agnostic.
 */
object HttpClientFactory {
    val json: Json = Json {
        ignoreUnknownKeys = true       // backend may add fields; clients shouldn't break
        explicitNulls = false           // omit nulls on the wire
        isLenient = true
        encodeDefaults = true
    }

    fun create(
        engine: HttpClientEngine,
        config: AppConfig,
        extraConfig: HttpClientConfig<*>.() -> Unit = {},
    ): HttpClient = HttpClient(engine) {
        expectSuccess = false  // we want to inspect non-2xx ourselves

        install(ContentNegotiation) {
            json(json)
        }

        install(HttpTimeout) {
            requestTimeoutMillis = REQUEST_TIMEOUT_MS
            connectTimeoutMillis = CONNECT_TIMEOUT_MS
            socketTimeoutMillis = SOCKET_TIMEOUT_MS
        }

        if (config.isDebug) {
            install(Logging) {
                level = LogLevel.INFO
                logger = object : Logger {
                    override fun log(message: String) {
                        Napier.d(tag = "Http", message = message)
                    }
                }
            }
        }

        install(DefaultRequest) {
            url(config.apiRoot)
            contentType(ContentType.Application.Json)
            header(HttpHeaders.Accept, ContentType.Application.Json.toString())
            header("X-Client-Platform", "mobile")
        }

        extraConfig()
    }

    private const val REQUEST_TIMEOUT_MS = 30_000L
    private const val CONNECT_TIMEOUT_MS = 15_000L
    private const val SOCKET_TIMEOUT_MS = 30_000L
}
