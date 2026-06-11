package com.lssgoo.core.storage

import com.russhwolf.settings.Settings

/**
 * Persistent storage for auth tokens.
 *
 * Backed by [Settings] which maps to:
 *  - Android → `SharedPreferences` (consider migrating to EncryptedSharedPreferences for prod)
 *  - iOS     → `NSUserDefaults` (consider migrating to Keychain via `KeychainSettings`)
 *
 * The web frontend stores these in HttpOnly cookies; on mobile we have no HttpOnly,
 * so durable secure storage is the equivalent boundary. The interface lets us swap
 * implementations without touching call sites.
 */
interface TokenStorage {
    suspend fun accessToken(): String?
    suspend fun refreshToken(): String?
    suspend fun save(
        accessToken: String,
        refreshToken: String,
        accessTokenExpiresAt: String? = null,
        refreshTokenExpiresAt: String? = null,
    )
    suspend fun clear()
    suspend fun hasTokens(): Boolean = accessToken() != null
}

class SettingsTokenStorage(private val settings: Settings) : TokenStorage {

    override suspend fun accessToken(): String? =
        settings.readNonEmpty(KEY_ACCESS_TOKEN)

    override suspend fun refreshToken(): String? =
        settings.readNonEmpty(KEY_REFRESH_TOKEN)

    override suspend fun save(
        accessToken: String,
        refreshToken: String,
        accessTokenExpiresAt: String?,
        refreshTokenExpiresAt: String?,
    ) {
        settings.putString(KEY_ACCESS_TOKEN, accessToken)
        settings.putString(KEY_REFRESH_TOKEN, refreshToken)
        accessTokenExpiresAt?.let { settings.putString(KEY_ACCESS_EXPIRES_AT, it) }
        refreshTokenExpiresAt?.let { settings.putString(KEY_REFRESH_EXPIRES_AT, it) }
    }

    override suspend fun clear() {
        listOf(
            KEY_ACCESS_TOKEN, KEY_REFRESH_TOKEN,
            KEY_ACCESS_EXPIRES_AT, KEY_REFRESH_EXPIRES_AT,
        ).forEach(settings::remove)
    }

    private companion object {
        const val KEY_ACCESS_TOKEN = "auth.access_token"
        const val KEY_REFRESH_TOKEN = "auth.refresh_token"
        const val KEY_ACCESS_EXPIRES_AT = "auth.access_expires_at"
        const val KEY_REFRESH_EXPIRES_AT = "auth.refresh_expires_at"
    }
}

private fun Settings.readNonEmpty(key: String): String? =
    if (hasKey(key)) getString(key, "").takeIf { it.isNotEmpty() } else null
