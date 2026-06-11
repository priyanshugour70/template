package com.lssgoo.navigation

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.List
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Person
import androidx.compose.material.icons.filled.Settings
import cafe.adriel.voyager.core.screen.Screen
import cafe.adriel.voyager.navigator.tab.CurrentTab
import cafe.adriel.voyager.navigator.tab.LocalTabNavigator
import cafe.adriel.voyager.navigator.tab.Tab
import cafe.adriel.voyager.navigator.tab.TabNavigator
import cafe.adriel.voyager.navigator.tab.TabOptions
import com.lssgoo.feature.dashboard.presentation.DashboardScreen
import com.lssgoo.feature.settings.presentation.SettingsScreen

/**
 * Post-login shell with bottom-tab navigation. Mobile equivalent of the web's
 * sidebar — same destinations, different chrome.
 *
 * Each tab is independent: its back stack is preserved when switching, so the
 * user doesn't lose their place when popping over to Settings and back.
 */
class MainScaffoldScreen : Screen {

    @Composable
    override fun Content() {
        TabNavigator(HomeTab) {
            Scaffold(
                content = { inner ->
                    androidx.compose.foundation.layout.Box(
                        modifier = Modifier.fillMaxSize().padding(inner),
                    ) { CurrentTab() }
                },
                bottomBar = {
                    NavigationBar(containerColor = MaterialTheme.colorScheme.surface) {
                        TabBarItem(HomeTab)
                        TabBarItem(BillingTab)
                        TabBarItem(ProfileTab)
                        TabBarItem(SettingsTab)
                    }
                },
            )
        }
    }
}

@Composable
private fun androidx.compose.foundation.layout.RowScope.TabBarItem(tab: Tab) {
    val nav = LocalTabNavigator.current
    NavigationBarItem(
        selected = nav.current.key == tab.key,
        onClick = { nav.current = tab },
        icon = {
            tab.options.icon?.let { Icon(it, contentDescription = tab.options.title) }
        },
        label = { Text(tab.options.title) },
    )
}

// ----------------- Tabs -----------------

object HomeTab : Tab {
    override val options: TabOptions
        @Composable get() = TabOptions(index = 0u, title = "Home", icon = rememberVector(Icons.Default.Home))
    @Composable override fun Content() { DashboardScreen().Content() }
}

object BillingTab : Tab {
    override val options: TabOptions
        @Composable get() = TabOptions(index = 1u, title = "Billing", icon = rememberVector(Icons.AutoMirrored.Filled.List))
    @Composable override fun Content() { PlaceholderTab("Billing", "Wire feature/billing/ next.") }
}

object ProfileTab : Tab {
    override val options: TabOptions
        @Composable get() = TabOptions(index = 2u, title = "Profile", icon = rememberVector(Icons.Default.Person))
    @Composable override fun Content() { PlaceholderTab("Profile", "Wire feature/user/ next.") }
}

object SettingsTab : Tab {
    override val options: TabOptions
        @Composable get() = TabOptions(index = 3u, title = "Settings", icon = rememberVector(Icons.Default.Settings))
    @Composable override fun Content() { SettingsScreen().Content() }
}

@Composable
private fun rememberVector(icon: ImageVector): androidx.compose.ui.graphics.painter.Painter =
    androidx.compose.ui.graphics.vector.rememberVectorPainter(icon)

@Composable
private fun PlaceholderTab(title: String, hint: String) {
    com.lssgoo.core.presentation.components.ScreenScaffold {
        com.lssgoo.core.presentation.components.EmptyState(
            title = title,
            description = hint,
        )
    }
}
