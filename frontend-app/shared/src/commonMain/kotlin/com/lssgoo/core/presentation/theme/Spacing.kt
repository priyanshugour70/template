package com.lssgoo.core.presentation.theme

import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp

/**
 * Tailwind-style spacing scale, exposed via `LocalSpacing` so screens never
 * inline magic numbers. Each step ≈ 4dp; matches the web's `space-x-N` semantics.
 */
data class Spacing(
    val none: Dp = 0.dp,
    val xs: Dp = 4.dp,
    val sm: Dp = 8.dp,
    val md: Dp = 12.dp,
    val lg: Dp = 16.dp,
    val xl: Dp = 24.dp,
    val xxl: Dp = 32.dp,
    val xxxl: Dp = 48.dp,
)

val LocalSpacing = staticCompositionLocalOf { Spacing() }
