package com.lssgoo.core.presentation.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.ColorScheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import org.koin.compose.koinInject

/**
 * Root theme wrapper. Resolves:
 *  - Active [Palette] (from [PaletteController], persisted)
 *  - Light vs dark variant (from user preference, fallback to system)
 *  - Material 3 [ColorScheme] derived from the palette tokens
 *  - [LocalSpacing], [LocalPalette], [LocalPaletteController] for downstream consumers
 *
 * Wrap the navigator root in this exactly once.
 */
val LocalPalette = staticCompositionLocalOf<PaletteTokens> {
    error("PaletteTokens not provided. Wrap composables in AppTheme { ... }")
}

@Composable
fun AppTheme(
    controller: PaletteController = koinInject(),
    content: @Composable () -> Unit,
) {
    val palette by controller.palette.collectAsState()
    val darkPref by controller.darkMode.collectAsState()
    val systemDark = isSystemInDarkTheme()
    val useDark = when (darkPref) {
        PaletteController.DarkModePreference.System -> systemDark
        PaletteController.DarkModePreference.Light -> false
        PaletteController.DarkModePreference.Dark -> true
    }
    val tokens = if (useDark) palette.dark else palette.light
    val colorScheme = tokens.toMaterialColorScheme(isDark = useDark)

    CompositionLocalProvider(
        LocalPaletteController provides controller,
        LocalPalette provides tokens,
        LocalSpacing provides Spacing(),
    ) {
        MaterialTheme(
            colorScheme = colorScheme,
            typography = AppTypography,
            shapes = AppShapes,
            content = content,
        )
    }
}

private fun PaletteTokens.toMaterialColorScheme(isDark: Boolean): ColorScheme {
    val base = if (isDark) darkColorScheme() else lightColorScheme()
    return base.copy(
        primary = primary,
        onPrimary = primaryForeground,
        primaryContainer = primary,
        onPrimaryContainer = primaryForeground,
        secondary = secondary,
        onSecondary = secondaryForeground,
        secondaryContainer = secondary,
        onSecondaryContainer = secondaryForeground,
        tertiary = accent,
        onTertiary = accentForeground,
        background = background,
        onBackground = foreground,
        surface = card,
        onSurface = cardForeground,
        surfaceVariant = muted,
        onSurfaceVariant = mutedForeground,
        error = destructive,
        onError = destructiveForeground,
        outline = border,
        outlineVariant = border,
    )
}
