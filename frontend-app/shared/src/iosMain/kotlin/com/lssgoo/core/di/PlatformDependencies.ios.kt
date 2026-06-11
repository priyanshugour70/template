package com.lssgoo.core.di

import com.russhwolf.settings.NSUserDefaultsSettings
import com.russhwolf.settings.Settings
import io.ktor.client.engine.HttpClientEngine
import io.ktor.client.engine.darwin.Darwin
import platform.Foundation.NSUserDefaults

/**
 * iOS: Darwin URLSession engine + NSUserDefaults-backed Settings.
 *
 * Note: NSUserDefaults is NOT secure storage. For production, switch to
 * `KeychainSettings` from `multiplatform-settings-keychain` once tokens go live.
 * Kept here for now to skip the Keychain entitlement dance during initial setup.
 */
actual fun provideHttpClientEngine(): HttpClientEngine = Darwin.create {
    configureRequest {
        setAllowsCellularAccess(true)
    }
}

actual fun provideSettings(): Settings =
    NSUserDefaultsSettings(NSUserDefaults.standardUserDefaults)
