package com.lssgoo.feature.settings.di

import com.lssgoo.feature.settings.presentation.SettingsScreenModel
import org.koin.core.module.dsl.factoryOf
import org.koin.dsl.module

val settingsModule = module {
    factoryOf(::SettingsScreenModel)
    // Add notification prefs, security, sessions, tenant settings here.
}
