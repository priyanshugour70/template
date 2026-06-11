package com.lssgoo.feature.auth.di

import com.lssgoo.feature.auth.data.AuthApi
import com.lssgoo.feature.auth.data.AuthRepositoryImpl
import com.lssgoo.feature.auth.domain.AuthRepository
import com.lssgoo.feature.auth.domain.usecase.BootstrapSessionUseCase
import com.lssgoo.feature.auth.domain.usecase.DiscoverTenantsUseCase
import com.lssgoo.feature.auth.domain.usecase.LoginUseCase
import com.lssgoo.feature.auth.domain.usecase.LogoutUseCase
import com.lssgoo.feature.auth.presentation.LoginScreenModel
import org.koin.core.module.dsl.factoryOf
import org.koin.dsl.module

val authModule = module {
    // Data layer
    single { AuthApi(client = get()) }
    single<AuthRepository> {
        AuthRepositoryImpl(
            api = get(),
            tokenStorage = get(),
            sessionCache = get(),
        )
    }

    // Use cases — factory because they're cheap and stateless
    factoryOf(::DiscoverTenantsUseCase)
    factoryOf(::LoginUseCase)
    factoryOf(::LogoutUseCase)
    factoryOf(::BootstrapSessionUseCase)

    // Screen models — factory: a new one per navigation entry
    factoryOf(::LoginScreenModel)
}
