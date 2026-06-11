package com.lssgoo.feature.billing.di

import org.koin.dsl.module

val billingModule = module {
    // single { BillingApi(client = get()) }
    // single<BillingRepository> { BillingRepositoryImpl(api = get()) }
    // factoryOf(::GetInvoicesUseCase)
    // factoryOf(::GetSubscriptionUseCase)
    // factoryOf(::BillingScreenModel)
}
