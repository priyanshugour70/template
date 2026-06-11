package com.lssgoo.feature.auth.presentation

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import cafe.adriel.voyager.core.screen.Screen
import cafe.adriel.voyager.koin.koinScreenModel
import cafe.adriel.voyager.navigator.LocalNavigator
import cafe.adriel.voyager.navigator.currentOrThrow
import com.lssgoo.core.presentation.components.EmptyState
import com.lssgoo.core.presentation.components.PrimaryButton
import com.lssgoo.core.presentation.components.PrimaryTextField
import com.lssgoo.core.presentation.components.ScreenScaffold
import com.lssgoo.core.presentation.theme.LocalSpacing
import com.lssgoo.feature.auth.domain.model.AuthTenant
import com.lssgoo.navigation.RootSwitchEvent
import com.lssgoo.navigation.RootEvents

/**
 * Multi-step login. Renders one of three sub-screens based on [LoginUiState.step].
 *
 * Navigation: on success, [LoginScreenModel] emits [LoginEvent.NavigateHome]; we
 * forward it to [RootEvents] which the [RootScreen] listens to. This avoids
 * coupling a feature screen to the global Navigator stack.
 */
class LoginScreen : Screen {

    @Composable
    override fun Content() {
        val model: LoginScreenModel = koinScreenModel()
        val state by model.state.collectAsState()

        LaunchedEffect(model) {
            model.events.collect { event ->
                when (event) {
                    LoginEvent.NavigateHome -> RootEvents.emit(RootSwitchEvent.NavigateHome)
                }
            }
        }

        ScreenScaffold {
            when (state.step) {
                LoginStep.Email -> EmailStep(state, model)
                LoginStep.TenantPicker -> TenantPickerStep(state, model)
                LoginStep.Password -> PasswordStep(state, model)
            }
        }
    }
}

@Composable
private fun EmailStep(state: LoginUiState, model: LoginScreenModel) {
    val spacing = LocalSpacing.current
    Column(
        modifier = Modifier.fillMaxSize().padding(top = spacing.xxxl),
        verticalArrangement = Arrangement.spacedBy(spacing.md),
    ) {
        Text("Welcome back", style = MaterialTheme.typography.displaySmall)
        Text(
            "Sign in to continue to your workspace.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Spacer(Modifier.height(spacing.lg))
        PrimaryTextField(
            value = state.email,
            onValueChange = model::onEmailChange,
            label = "Email",
            keyboardType = KeyboardType.Email,
            error = state.error,
            enabled = !state.isLoading,
        )
        Spacer(Modifier.height(spacing.sm))
        PrimaryButton(
            text = "Continue",
            onClick = model::onContinueEmail,
            loading = state.isLoading,
        )
    }
}

@Composable
private fun TenantPickerStep(state: LoginUiState, model: LoginScreenModel) {
    val spacing = LocalSpacing.current
    Column(
        modifier = Modifier.fillMaxSize().padding(top = spacing.xxxl),
        verticalArrangement = Arrangement.spacedBy(spacing.md),
    ) {
        Text("Pick a workspace", style = MaterialTheme.typography.headlineMedium)
        Text(
            "${state.email} belongs to ${state.tenants.size} workspaces.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Spacer(Modifier.height(spacing.sm))
        if (state.tenants.isEmpty()) {
            EmptyState(
                title = "No workspaces found",
                description = "Ask your admin to invite you.",
                actionLabel = "Back",
                onAction = model::onBackToEmail,
            )
        } else {
            LazyColumn(verticalArrangement = Arrangement.spacedBy(spacing.sm)) {
                items(state.tenants, key = AuthTenant::id) { tenant ->
                    TenantRow(tenant) { model.onSelectTenant(tenant) }
                }
            }
        }
        Spacer(Modifier.height(spacing.md))
        TextButton(onClick = model::onBackToEmail) {
            Text("Use a different email")
        }
    }
}

@Composable
private fun TenantRow(tenant: AuthTenant, onClick: () -> Unit) {
    val spacing = LocalSpacing.current
    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        onClick = onClick,
    ) {
        Column(
            modifier = Modifier.fillMaxWidth().padding(spacing.lg),
            verticalArrangement = Arrangement.spacedBy(spacing.xs),
        ) {
            Text(tenant.name, style = MaterialTheme.typography.titleMedium)
            tenant.slug?.let {
                Text(it, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
        }
    }
}

@Composable
private fun PasswordStep(state: LoginUiState, model: LoginScreenModel) {
    val spacing = LocalSpacing.current
    Column(
        modifier = Modifier.fillMaxSize().padding(top = spacing.xxxl),
        verticalArrangement = Arrangement.spacedBy(spacing.md),
    ) {
        Text("Enter your password", style = MaterialTheme.typography.headlineMedium)
        Text(
            state.selectedTenant?.let { "Signing in to ${it.name} as ${state.email}" }
                ?: "Signing in as ${state.email}",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Spacer(Modifier.height(spacing.lg))
        PrimaryTextField(
            value = state.password,
            onValueChange = model::onPasswordChange,
            label = "Password",
            keyboardType = KeyboardType.Password,
            error = state.error,
            enabled = !state.isLoading,
        )
        Spacer(Modifier.height(spacing.sm))
        PrimaryButton(
            text = "Sign In",
            onClick = model::onSubmitPassword,
            loading = state.isLoading,
        )
        TextButton(
            onClick = model::onBackToEmail,
            modifier = Modifier.align(Alignment.CenterHorizontally),
        ) {
            Text("Use a different email")
        }
    }
}
