package com.lssgoo.core.storage

import com.lssgoo.feature.auth.domain.model.Session
import com.russhwolf.settings.Settings
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.serialization.SerializationException
import kotlinx.serialization.json.Json

/**
 * Mirrors the web's `useSessionStore` (Zustand) — fast, durable cache of the
 * last-authenticated session so the UI shell can render before the network confirms.
 *
 * Two responsibilities:
 *  - Persist [Session] JSON in [Settings] across app launches.
 *  - Expose a [StateFlow] so Composables can collect updates reactively.
 *
 * On first app launch this returns null; after the first successful login it
 * survives kills. Cleared on logout.
 */
class SessionCache(
    private val settings: Settings,
    private val json: Json,
) {
    private val state = MutableStateFlow<Session?>(loadFromDisk())
    val session: StateFlow<Session?> = state.asStateFlow()

    val current: Session? get() = state.value

    fun update(session: Session?) {
        state.value = session
        if (session == null) {
            settings.remove(KEY_SESSION)
        } else {
            settings.putString(KEY_SESSION, json.encodeToString(Session.serializer(), session))
        }
    }

    fun clear() = update(null)

    private fun loadFromDisk(): Session? {
        if (!settings.hasKey(KEY_SESSION)) return null
        return try {
            json.decodeFromString(Session.serializer(), settings.getString(KEY_SESSION, ""))
        } catch (_: SerializationException) {
            settings.remove(KEY_SESSION)
            null
        }
    }

    private companion object {
        const val KEY_SESSION = "session.cached_v1"
    }
}
