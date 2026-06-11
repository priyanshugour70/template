package com.lssgoo.core.di

import com.lssgoo.core.config.AppConfig
import com.lssgoo.feature.auth.di.authModule
import com.lssgoo.feature.billing.di.billingModule
import com.lssgoo.feature.dashboard.di.dashboardModule
import com.lssgoo.feature.settings.di.settingsModule
import com.lssgoo.feature.user.di.userModule
import io.github.aakira.napier.DebugAntilog
import io.github.aakira.napier.Napier
import org.koin.core.KoinApplication
import org.koin.core.context.startKoin
import org.koin.dsl.KoinAppDeclaration

/**
 * The single entry point platforms call to bring the shared module to life.
 *
 *  - Android: call from `Application.onCreate`
 *  - iOS:     call from the app's `init` (Swift bridge)
 *
 * Order matters: Napier first so any error during module init is logged;
 * Koin second so the feature graph is ready before the UI mounts.
 */
fun initApp(config: AppConfig, extraConfig: KoinAppDeclaration? = null): KoinApplication {
    if (config.isDebug) {
        Napier.base(DebugAntilog())
    }
    return startKoin {
        extraConfig?.invoke(this)
        modules(
            coreModule(config),
            authModule,
            dashboardModule,
            billingModule,
            userModule,
            settingsModule,
        )
    }
}
