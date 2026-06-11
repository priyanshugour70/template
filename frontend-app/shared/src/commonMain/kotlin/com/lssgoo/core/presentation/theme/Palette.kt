package com.lssgoo.core.presentation.theme

import androidx.compose.ui.graphics.Color

/**
 * Mirrors the web's `ThemeTokens` interface from `src/theme/palettes.ts`. Same field
 * names, same intent. Changes to either side should be mirrored to keep design parity.
 *
 * We use Compose [Color] rather than hex strings because:
 *  - It's the type [MaterialTheme]/[ColorScheme] expects natively.
 *  - It eliminates an entire class of runtime parsing errors.
 */
data class PaletteTokens(
    // Surfaces
    val background: Color,
    val card: Color,
    val popover: Color,
    val muted: Color,
    val border: Color,
    val input: Color,
    val ring: Color,
    // Text
    val foreground: Color,
    val cardForeground: Color,
    val popoverForeground: Color,
    val mutedForeground: Color,
    // Brand roles
    val primary: Color,
    val primaryForeground: Color,
    val secondary: Color,
    val secondaryForeground: Color,
    val accent: Color,
    val accentForeground: Color,
    val success: Color,
    val successForeground: Color,
    val destructive: Color,
    val destructiveForeground: Color,
)

/**
 * A palette has a light + dark variant. The active variant is chosen by [AppTheme]
 * based on system dark-mode and the user's override preference.
 */
data class Palette(
    val id: PaletteId,
    val displayName: String,
    val light: PaletteTokens,
    val dark: PaletteTokens,
)

enum class PaletteId(val raw: String) {
    ForestTrail("forest-trail"),
    SunsetHorizon("sunset-horizon"),
    TropicalParadise("tropical-paradise"),
    MountainMist("mountain-mist"),
    DesertDunes("desert-dunes");

    companion object {
        fun fromRaw(raw: String?): PaletteId =
            entries.firstOrNull { it.raw == raw } ?: ForestTrail
    }
}
