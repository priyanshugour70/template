package com.lssgoo.feature.dashboard.di

import org.koin.dsl.module

/**
 * Empty for now. When you add the dashboard's repository / use cases / screen
 * model, register them here. The module is included from
 * [com.lssgoo.core.di.initApp] so anything you add becomes available immediately.
 */
val dashboardModule = module {
    // single { DashboardApi(client = get()) }
    // single<DashboardRepository> { DashboardRepositoryImpl(api = get()) }
    // factoryOf(::GetDashboardSummaryUseCase)
    // factoryOf(::DashboardScreenModel)
}
