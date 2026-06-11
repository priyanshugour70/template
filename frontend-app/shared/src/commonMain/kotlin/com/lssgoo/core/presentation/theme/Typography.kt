package com.lssgoo.core.presentation.theme

import androidx.compose.material3.Typography
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp

/**
 * The web uses Inter (`--font-inter`). Compose Multiplatform can't bundle Google
 * Fonts portably without per-platform setup, so we default to the platform's
 * sans-serif here (San Francisco on iOS, Roboto on Android). When/if we add the
 * Inter `.otf` to `composeResources/font/`, swap `FontFamily.SansSerif` for the
 * loaded family via [androidx.compose.ui.text.font.Font].
 *
 * Size scale matches Tailwind's `text-*` defaults.
 */
val InterLikeFont: FontFamily = FontFamily.SansSerif

val AppTypography = Typography(
    displayLarge = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.SemiBold, fontSize = 48.sp, lineHeight = 56.sp),
    displayMedium = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.SemiBold, fontSize = 36.sp, lineHeight = 44.sp),
    displaySmall = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.SemiBold, fontSize = 30.sp, lineHeight = 38.sp),

    headlineLarge = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.SemiBold, fontSize = 24.sp, lineHeight = 32.sp),
    headlineMedium = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.SemiBold, fontSize = 20.sp, lineHeight = 28.sp),
    headlineSmall = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.SemiBold, fontSize = 18.sp, lineHeight = 26.sp),

    titleLarge = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.Medium, fontSize = 18.sp, lineHeight = 24.sp),
    titleMedium = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.Medium, fontSize = 16.sp, lineHeight = 22.sp),
    titleSmall = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.Medium, fontSize = 14.sp, lineHeight = 20.sp),

    bodyLarge = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.Normal, fontSize = 16.sp, lineHeight = 24.sp),
    bodyMedium = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.Normal, fontSize = 14.sp, lineHeight = 20.sp),
    bodySmall = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.Normal, fontSize = 12.sp, lineHeight = 16.sp),

    labelLarge = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.Medium, fontSize = 14.sp, lineHeight = 20.sp),
    labelMedium = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.Medium, fontSize = 12.sp, lineHeight = 16.sp),
    labelSmall = TextStyle(fontFamily = InterLikeFont, fontWeight = FontWeight.Medium, fontSize = 11.sp, lineHeight = 14.sp),
)
