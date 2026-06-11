package com.lssgoo.feature.dashboard.presentation

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import cafe.adriel.voyager.core.screen.Screen
import com.lssgoo.core.presentation.components.PrimaryButton
import com.lssgoo.core.presentation.components.ScreenScaffold
import com.lssgoo.core.presentation.theme.LocalSpacing
import com.lssgoo.feature.auth.domain.AuthRepository
import com.lssgoo.feature.auth.domain.usecase.LogoutUseCase
import com.lssgoo.navigation.RootEvents
import com.lssgoo.navigation.RootSwitchEvent
import kotlinx.coroutines.launch
import org.koin.compose.koinInject

/**
 * Placeholder dashboard. Replace with real analytics tiles (matching the web's
 * `(dashboard)/dashboard/page.tsx`) — wire a `DashboardScreenModel` + repository
 * + `GetDashboardSummaryUseCase` following the auth feature as a template.
 */
class DashboardScreen : Screen {

    @Composable
    override fun Content() {
        val spacing = LocalSpacing.current
        val authRepo: AuthRepository = koinInject()
        val logout: LogoutUseCase = koinInject()
        val session by authRepo.session.collectAsState()
        val scope = rememberCoroutineScope()

        ScreenScaffold {
            Column(
                modifier = Modifier.fillMaxSize().padding(top = spacing.xxxl),
                verticalArrangement = Arrangement.spacedBy(spacing.md),
            ) {
                Text(
                    text = "Welcome${session?.user?.firstName?.let { ", $it" } ?: ""}",
                    style = MaterialTheme.typography.displaySmall,
                )
                session?.tenant?.let {
                    Text(
                        text = "Workspace: ${it.name}",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }
                Text(
                    text = "Dashboard content goes here. Replace this with real tiles.",
                    style = MaterialTheme.typography.bodyMedium,
                )
                PrimaryButton(
                    text = "Sign out",
                    onClick = {
                        scope.launch {
                            logout()
                            RootEvents.emit(RootSwitchEvent.NavigateLogin)
                        }
                    },
                )
            }
        }
    }
}

@Composable
private fun rememberCoroutineScope() = androidx.compose.runtime.rememberCoroutineScope()
