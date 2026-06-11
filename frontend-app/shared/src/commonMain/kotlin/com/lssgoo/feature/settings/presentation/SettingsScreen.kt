package com.lssgoo.feature.settings.presentation

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import cafe.adriel.voyager.core.screen.Screen
import cafe.adriel.voyager.koin.koinScreenModel
import com.lssgoo.core.presentation.components.ScreenScaffold
import com.lssgoo.core.presentation.theme.LocalSpacing
import com.lssgoo.core.presentation.theme.PaletteController
import com.lssgoo.core.presentation.theme.Palettes

/**
 * Settings shell. Today: palette + dark mode. Tomorrow: notifications, security,
 * sessions, tenant settings, etc. Each section becomes its own composable here
 * with its own use case wired through the [SettingsScreenModel].
 */
class SettingsScreen : Screen {

    @Composable
    override fun Content() {
        val model: SettingsScreenModel = koinScreenModel()
        val state by model.state.collectAsState()
        val spacing = LocalSpacing.current

        ScreenScaffold {
            Column(
                modifier = Modifier.fillMaxSize().padding(top = spacing.xxxl),
                verticalArrangement = Arrangement.spacedBy(spacing.lg),
            ) {
                Text("Settings", style = MaterialTheme.typography.displaySmall)

                Text("Theme palette", style = MaterialTheme.typography.titleMedium)
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(spacing.sm),
                ) {
                    Palettes.all.forEach { p ->
                        FilterChip(
                            selected = state.currentPalette == p.id,
                            onClick = { model.selectPalette(p.id) },
                            label = { Text(p.displayName) },
                        )
                    }
                }

                Text("Appearance", style = MaterialTheme.typography.titleMedium)
                Row(horizontalArrangement = Arrangement.spacedBy(spacing.sm)) {
                    PaletteController.DarkModePreference.entries.forEach { pref ->
                        FilterChip(
                            selected = state.darkMode == pref,
                            onClick = { model.selectDarkMode(pref) },
                            label = { Text(pref.name) },
                        )
                    }
                }
            }
        }
    }
}
