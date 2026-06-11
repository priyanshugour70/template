package com.lssgoo.feature.auth.presentation

import com.lssgoo.core.presentation.BaseScreenModel
import com.lssgoo.core.result.AppResult
import com.lssgoo.feature.auth.domain.model.AuthTenant
import com.lssgoo.feature.auth.domain.usecase.DiscoverTenantsUseCase
import com.lssgoo.feature.auth.domain.usecase.LoginUseCase

/**
 * MVI state machine for the multi-step login screen.
 *
 *   Step.Email     → user types email, taps Continue → discoverTenants
 *      └ 0 tenants → Step.Password with tenantId=null (signup-via-invite, etc.)
 *      └ 1 tenant  → Step.Password with that tenantId auto-selected
 *      └ N tenants → Step.TenantPicker
 *   Step.TenantPicker → user picks → Step.Password
 *   Step.Password  → user types pw, taps Sign In → login → emit NavigateHome
 *
 * Errors are surfaced via [LoginUiState.error]. Side effects (navigation) flow
 * through [LoginEvent].
 */
class LoginScreenModel(
    private val discoverTenants: DiscoverTenantsUseCase,
    private val login: LoginUseCase,
) : BaseScreenModel<LoginUiState, LoginEvent>(LoginUiState()) {

    fun onEmailChange(value: String) = updateState { copy(email = value, error = null) }
    fun onPasswordChange(value: String) = updateState { copy(password = value, error = null) }
    fun onSelectTenant(tenant: AuthTenant) {
        updateState { copy(selectedTenant = tenant, step = LoginStep.Password, error = null) }
    }

    fun onBackToEmail() = updateState {
        copy(step = LoginStep.Email, password = "", selectedTenant = null, tenants = emptyList(), error = null)
    }

    fun onContinueEmail() = launch {
        val current = state.value
        if (current.email.isBlank()) {
            updateState { copy(error = "Please enter your email.") }
            return@launch
        }
        updateState { copy(isLoading = true, error = null) }
        when (val result = discoverTenants(current.email)) {
            is AppResult.Success -> {
                val tenants = result.data
                updateState {
                    copy(
                        isLoading = false,
                        tenants = tenants,
                        step = when {
                            tenants.isEmpty() -> LoginStep.Password
                            tenants.size == 1 -> LoginStep.Password
                            else -> LoginStep.TenantPicker
                        },
                        selectedTenant = tenants.singleOrNull(),
                    )
                }
            }
            is AppResult.Failure -> updateState {
                copy(isLoading = false, error = result.error.message)
            }
        }
    }

    fun onSubmitPassword() = launch {
        val current = state.value
        if (current.password.isBlank()) {
            updateState { copy(error = "Please enter your password.") }
            return@launch
        }
        updateState { copy(isLoading = true, error = null) }
        val result = login(
            LoginUseCase.Params(
                email = current.email,
                password = current.password,
                tenantId = current.selectedTenant?.id,
            ),
        )
        when (result) {
            is AppResult.Success -> {
                updateState { copy(isLoading = false) }
                emitEvent(LoginEvent.NavigateHome)
            }
            is AppResult.Failure -> updateState {
                copy(isLoading = false, error = result.error.message)
            }
        }
    }
}

enum class LoginStep { Email, TenantPicker, Password }

data class LoginUiState(
    val step: LoginStep = LoginStep.Email,
    val email: String = "",
    val password: String = "",
    val tenants: List<AuthTenant> = emptyList(),
    val selectedTenant: AuthTenant? = null,
    val isLoading: Boolean = false,
    val error: String? = null,
)

sealed interface LoginEvent {
    data object NavigateHome : LoginEvent
}
