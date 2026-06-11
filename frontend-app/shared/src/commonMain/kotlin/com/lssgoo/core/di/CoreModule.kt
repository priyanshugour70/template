package com.lssgoo.core.di

import com.lssgoo.core.config.AppConfig
import com.lssgoo.core.network.ApiClient
import com.lssgoo.core.network.HttpClientFactory
import com.lssgoo.core.network.TokenRefresher
import com.lssgoo.core.presentation.theme.PaletteController
import com.lssgoo.core.storage.SessionCache
import com.lssgoo.core.storage.SettingsTokenStorage
import com.lssgoo.core.storage.TokenStorage
import io.ktor.client.HttpClient
import kotlinx.serialization.json.Json
import org.koin.core.qualifier.named
import org.koin.dsl.module

/**
 * Wires every cross-cutting dependency exactly once.
 *
 * Convention: features depend on interfaces declared here (e.g. [ApiClient],
 * [TokenStorage]) and never reach for [HttpClient] directly. That keeps the
 * refresh-on-401 logic in one place and makes feature code trivial to test
 * with fakes.
 */
fun coreModule(config: AppConfig) = module {

    // --- Config ---------------------------------------------------------
    single { config }
    single { HttpClientFactory.json }

    // --- Platform deps --------------------------------------------------
    single { provideSettings() }
    single { provideHttpClientEngine() }

    // --- Network --------------------------------------------------------
    single<HttpClient> {
        HttpClientFactory.create(engine = get(), config = get())
    }
    single { TokenRefresher(httpClient = get(), tokenStorage = get()) }
    single { ApiClient(client = get(), tokenStorage = get(), refresher = get()) }

    // --- Storage --------------------------------------------------------
    single<TokenStorage> { SettingsTokenStorage(settings = get()) }
    single { SessionCache(settings = get(), json = get<Json>()) }

    // --- Theme ----------------------------------------------------------
    single { PaletteController(settings = get()) }
}
