package com.lssgoo

import androidx.compose.ui.window.ComposeUIViewController
import com.lssgoo.core.config.AppConfig
import com.lssgoo.core.di.initApp
import platform.UIKit.UIViewController

/**
 * Bridge from SwiftUI to Compose Multiplatform.
 *
 * Swift calls this once at app launch via:
 * ```swift
 * MainViewControllerKt.MainViewController(apiBaseUrl: "https://api.example.com")
 * ```
 *
 * Side effect: starts Koin on the first invocation. Safe to call multiple times
 * (idempotent guard via [koinStarted]).
 */
private var koinStarted = false

fun MainViewController(apiBaseUrl: String = "http://localhost:8080"): UIViewController {
    if (!koinStarted) {
        initApp(
            config = AppConfig(
                apiBaseUrl = apiBaseUrl,
                isDebug = true,
                environment = "development",
            ),
        )
        koinStarted = true
    }
    return ComposeUIViewController { App() }
}
