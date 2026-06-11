package com.lssgoo.feature.settings.presentation

import com.lssgoo.core.presentation.BaseScreenModel
import com.lssgoo.core.presentation.theme.PaletteController
import com.lssgoo.core.presentation.theme.PaletteId

/**
 * The only real settings wired in this scaffold: palette + dark mode. Wire the
 * rest (notifications, security, sessions) the same way — inject a use case or
 * repository, expose state, mutate via methods.
 */
class SettingsScreenModel(
    private val paletteController: PaletteController,
) : BaseScreenModel<SettingsUiState, Unit>(SettingsUiState()) {

    init {
        launch {
            paletteController.palette.collect { p ->
                updateState { copy(currentPalette = p.id) }
            }
        }
        launch {
            paletteController.darkMode.collect { pref ->
                updateState { copy(darkMode = pref) }
            }
        }
    }

    fun selectPalette(id: PaletteId) = paletteController.selectPalette(id)
    fun selectDarkMode(pref: PaletteController.DarkModePreference) =
        paletteController.selectDarkMode(pref)
}

data class SettingsUiState(
    val currentPalette: PaletteId = PaletteId.ForestTrail,
    val darkMode: PaletteController.DarkModePreference = PaletteController.DarkModePreference.System,
)
