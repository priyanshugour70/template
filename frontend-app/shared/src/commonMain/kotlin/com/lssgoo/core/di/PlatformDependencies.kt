package com.lssgoo.core.di

import com.russhwolf.settings.Settings
import io.ktor.client.engine.HttpClientEngine

/**
 * Platform-specific dependencies we need but can't construct in `commonMain`.
 * The `expect` declarations are fulfilled by:
 *  - `androidMain` → OkHttp engine + Android SharedPreferences-backed Settings
 *  - `iosMain`     → Darwin engine + NSUserDefaults-backed Settings
 *
 * Kept in a single file so adding a new platform target stays a one-stop edit.
 */
expect fun provideHttpClientEngine(): HttpClientEngine
expect fun provideSettings(): Settings
