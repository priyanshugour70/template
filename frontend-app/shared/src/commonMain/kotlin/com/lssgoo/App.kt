package com.lssgoo

import androidx.compose.runtime.Composable
import cafe.adriel.voyager.navigator.Navigator
import com.lssgoo.navigation.RootScreen

/**
 * Top-level composable. Mounted by:
 *  - Android: `MainActivity.setContent { App() }`
 *  - iOS:     `MainViewController()` (which wraps this in a ComposeUIViewController)
 *
 * Everything happens under [RootScreen] — theme, DI, auth gating, navigation.
 */
@Composable
fun App() {
    Navigator(RootScreen())
}
