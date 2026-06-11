package com.lssgoo.feature.auth.domain.usecase

import com.lssgoo.core.domain.UseCase
import com.lssgoo.core.result.AppError
import com.lssgoo.core.result.AppResult
import com.lssgoo.core.result.ErrorCode
import com.lssgoo.feature.auth.domain.AuthRepository
import com.lssgoo.feature.auth.domain.model.AuthTenant

/**
 * Step 1 of the three-step login flow:
 * "given an email, which workspaces does this user belong to?"
 *
 * Input validation lives here — repositories shouldn't second-guess their inputs.
 */
class DiscoverTenantsUseCase(
    private val repository: AuthRepository,
) : UseCase<String, List<AuthTenant>> {

    override suspend fun invoke(params: String): AppResult<List<AuthTenant>> {
        val email = params.trim().lowercase()
        if (!email.isValidEmailLike()) {
            return AppResult.Failure(
                AppError(code = ErrorCode.VALIDATION, message = "Please enter a valid email."),
            )
        }
        return repository.discoverTenants(email)
    }

    private fun String.isValidEmailLike(): Boolean {
        // Simple structural check — server is the source of truth for real validation.
        val parts = split("@")
        if (parts.size != 2) return false
        val (local, domain) = parts
        if (local.isBlank() || domain.isBlank()) return false
        return "." in domain
    }
}
