package com.lssgoo.feature.auth.domain.usecase

import com.lssgoo.core.domain.UseCase
import com.lssgoo.core.result.AppError
import com.lssgoo.core.result.AppResult
import com.lssgoo.core.result.ErrorCode
import com.lssgoo.feature.auth.domain.AuthRepository
import com.lssgoo.feature.auth.domain.model.Session

/**
 * Step 2 of the three-step login flow: authenticate with credentials + chosen tenant.
 */
class LoginUseCase(
    private val repository: AuthRepository,
) : UseCase<LoginUseCase.Params, Session> {

    data class Params(
        val email: String,
        val password: String,
        val tenantId: String?,
    )

    override suspend fun invoke(params: Params): AppResult<Session> {
        if (params.email.isBlank() || params.password.isBlank()) {
            return AppResult.Failure(
                AppError(code = ErrorCode.VALIDATION, message = "Email and password are required."),
            )
        }
        if (params.password.length < 6) {
            return AppResult.Failure(
                AppError(code = ErrorCode.VALIDATION, message = "Password must be at least 6 characters."),
            )
        }
        return repository.login(
            email = params.email.trim().lowercase(),
            password = params.password,
            tenantId = params.tenantId,
        )
    }
}
