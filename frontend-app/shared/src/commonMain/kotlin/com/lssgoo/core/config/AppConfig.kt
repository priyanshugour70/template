package com.lssgoo.core.config

/**
 * Static, build-time configuration. Platform code (Android `Application`, iOS app delegate)
 * decides the value and hands it to [com.lssgoo.core.di.initKoin]. We deliberately do NOT
 * use BuildKonfig / BuildConfig here so the shared module stays plugin-light and a single
 * app binary can target multiple backends by swapping this object at startup.
 */
data class AppConfig(
    /** Backend root, e.g. `https://api.example.com`. No trailing slash. */
    val apiBaseUrl: String,
    /** Mounted under [apiBaseUrl]. Mirrors the Go backend's `/api/v1` prefix. */
    val apiPathPrefix: String = "/api/v1",
    val isDebug: Boolean = false,
    /** Free-form tag surfaced in logs and crash reports. */
    val environment: String = "development",
) {
    val apiRoot: String get() = "$apiBaseUrl$apiPathPrefix"
}
