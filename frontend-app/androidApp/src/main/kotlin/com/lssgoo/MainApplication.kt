package com.lssgoo

import android.app.Application
import com.lssgoo.core.config.AppConfig
import com.lssgoo.core.di.AppInitializer
import com.lssgoo.core.di.initApp
import org.koin.android.ext.koin.androidContext
import org.koin.android.ext.koin.androidLogger

/**
 * Entry point for the Android process. Owns one job: hand the shared module its
 * platform dependencies, then start Koin.
 *
 * Add this to `AndroidManifest.xml` with `android:name=".MainApplication"` so the
 * runtime instantiates it before any Activity.
 */
class MainApplication : Application() {

    override fun onCreate() {
        super.onCreate()

        // Surface the application context to the shared module BEFORE Koin starts
        // — provideSettings() needs it.
        AppInitializer.setup(this)

        initApp(
            config = AppConfig(
                // TODO: For real builds, switch this on BuildConfig.DEBUG or a
                // gradle build variant. Hard-coded today so the scaffold runs
                // out of the box against the local Go backend (`backend/`).
                apiBaseUrl = "http://10.0.2.2:8080",  // 10.0.2.2 = host loopback from Android emulator
                isDebug = true,
                environment = "development",
            ),
        ) {
            androidLogger()
            androidContext(this@MainApplication)
        }
    }
}
