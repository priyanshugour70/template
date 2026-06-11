package com.lssgoo.core.di

import android.content.Context
import com.russhwolf.settings.Settings
import com.russhwolf.settings.SharedPreferencesSettings
import io.ktor.client.engine.HttpClientEngine
import io.ktor.client.engine.okhttp.OkHttp

/**
 * Android: SharedPreferences-backed Settings + OkHttp engine.
 *
 * `appContext` is initialised by [AppInitializer] which the app's `Application`
 * calls in `onCreate`. We keep this as a process-level static because Koin's
 * Android module loads after our shared `initApp` and we need the value earlier.
 */
private lateinit var appContext: Context

object AppInitializer {
    fun setup(context: Context) {
        appContext = context.applicationContext
    }
}

actual fun provideHttpClientEngine(): HttpClientEngine = OkHttp.create {
    config {
        retryOnConnectionFailure(true)
    }
}

actual fun provideSettings(): Settings {
    check(::appContext.isInitialized) {
        "AppInitializer.setup(context) must be called from Application.onCreate before Koin starts."
    }
    val prefs = appContext.getSharedPreferences("lssgoo_app_prefs", Context.MODE_PRIVATE)
    return SharedPreferencesSettings(prefs)
}
