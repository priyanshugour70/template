package com.lssgoo.navigation

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import cafe.adriel.voyager.core.screen.Screen
import cafe.adriel.voyager.navigator.Navigator
import com.lssgoo.core.presentation.theme.AppTheme
import com.lssgoo.core.result.AppResult
import com.lssgoo.feature.auth.domain.usecase.BootstrapSessionUseCase
import com.lssgoo.feature.auth.presentation.LoginScreen
import org.koin.compose.koinInject

/**
 * The single screen mounted under [Navigator]. Decides between the auth flow and
 * the post-login scaffold based on:
 *  1. Initial bootstrap (call /auth/me with stored token) on first composition.
 *  2. [RootEvents] emissions from anywhere in the app (login success, logout).
 *
 * This is the equivalent of the web's `src/middleware.ts` + `(auth)` vs
 * `(dashboard)` route groups, collapsed into one place because mobile has no
 * URL bar to drive routing.
 */
class RootScreen : Screen {

    @Composable
    override fun Content() {
        AppTheme {
            val bootstrap: BootstrapSessionUseCase = koinInject()
            var phase: Phase by remember { mutableStateOf(Phase.Booting) }

            // Initial bootstrap on first composition.
            LaunchedEffect(Unit) {
                phase = when (bootstrap()) {
                    is AppResult.Success -> Phase.Main
                    is AppResult.Failure -> Phase.Auth
                }
            }

            // Listen for cross-feature switches.
            LaunchedEffect(Unit) {
                RootEvents.flow.collect { event ->
                    phase = when (event) {
                        RootSwitchEvent.NavigateHome -> Phase.Main
                        RootSwitchEvent.NavigateLogin -> Phase.Auth
                    }
                }
            }

            when (phase) {
                Phase.Booting -> BootingPlaceholder()
                Phase.Auth -> Navigator(LoginScreen())
                Phase.Main -> Navigator(MainScaffoldScreen())
            }
        }
    }

    private enum class Phase { Booting, Auth, Main }
}

@Composable
private fun BootingPlaceholder() {
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        CircularProgressIndicator()
    }
}
