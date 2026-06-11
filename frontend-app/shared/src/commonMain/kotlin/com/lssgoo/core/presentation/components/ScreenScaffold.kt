package com.lssgoo.core.presentation.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import com.lssgoo.core.presentation.theme.LocalSpacing

/**
 * Every screen should start with this. It applies:
 *  - The active palette's background color (avoids white-flash on theme switch)
 *  - Default horizontal padding (`lg` spacing token)
 *  - Material 3 [Scaffold] for snackbars / top bars when needed
 */
@Composable
fun ScreenScaffold(
    modifier: Modifier = Modifier,
    topBar: @Composable () -> Unit = {},
    bottomBar: @Composable () -> Unit = {},
    contentPadding: PaddingValues = PaddingValues(horizontal = LocalSpacing.current.lg),
    content: @Composable (PaddingValues) -> Unit,
) {
    Scaffold(
        modifier = modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background),
        containerColor = MaterialTheme.colorScheme.background,
        topBar = topBar,
        bottomBar = bottomBar,
    ) { inner ->
        Box(
            Modifier
                .padding(inner)
                .padding(contentPadding),
        ) {
            content(inner)
        }
    }
}
