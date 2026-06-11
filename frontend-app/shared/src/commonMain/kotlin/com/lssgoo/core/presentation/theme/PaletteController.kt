package com.lssgoo.core.presentation.theme

import androidx.compose.runtime.staticCompositionLocalOf
import com.russhwolf.settings.Settings
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

/**
 * Holds the active palette + dark-mode preference. Persists choices so they
 * survive process death — same model as the web's `usePaletteStore`.
 *
 * Mutations are synchronous on the UI thread; persistence is a fire-and-forget
 * write. We never re-read from disk after the constructor.
 */
class PaletteController(private val settings: Settings) {

    enum class DarkModePreference { System, Light, Dark }

    private val _palette = MutableStateFlow(loadPalette())
    val palette: StateFlow<Palette> = _palette.asStateFlow()

    private val _darkMode = MutableStateFlow(loadDarkMode())
    val darkMode: StateFlow<DarkModePreference> = _darkMode.asStateFlow()

    fun selectPalette(id: PaletteId) {
        _palette.value = Palettes.byId(id)
        settings.putString(KEY_PALETTE, id.raw)
    }

    fun selectDarkMode(pref: DarkModePreference) {
        _darkMode.value = pref
        settings.putString(KEY_DARK_MODE, pref.name)
    }

    private fun loadPalette(): Palette {
        val raw = if (settings.hasKey(KEY_PALETTE)) settings.getString(KEY_PALETTE, "") else ""
        return Palettes.byId(PaletteId.fromRaw(raw.ifBlank { null }))
    }

    private fun loadDarkMode(): DarkModePreference {
        val raw = if (settings.hasKey(KEY_DARK_MODE)) settings.getString(KEY_DARK_MODE, "") else ""
        return DarkModePreference.entries.firstOrNull { it.name == raw } ?: DarkModePreference.System
    }

    private companion object {
        const val KEY_PALETTE = "theme.palette_id"
        const val KEY_DARK_MODE = "theme.dark_mode"
    }
}

val LocalPaletteController = staticCompositionLocalOf<PaletteController> {
    error("PaletteController was not provided. Wrap your root composable in AppTheme { ... }")
}
