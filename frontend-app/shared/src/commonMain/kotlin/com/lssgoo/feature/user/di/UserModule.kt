package com.lssgoo.feature.user.di

import org.koin.dsl.module

val userModule = module {
    // single { UserApi(client = get()) }
    // single<UserRepository> { UserRepositoryImpl(api = get()) }
    // factoryOf(::GetCurrentUserUseCase)
    // factoryOf(::UpdateProfileUseCase)
    // factoryOf(::ProfileScreenModel)
}
