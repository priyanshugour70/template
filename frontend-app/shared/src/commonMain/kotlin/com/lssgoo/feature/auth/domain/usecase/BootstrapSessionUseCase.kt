package com.lssgoo.feature.auth.domain.usecase

import com.lssgoo.core.domain.NoParamUseCase
import com.lssgoo.core.result.AppResult
import com.lssgoo.feature.auth.domain.AuthRepository
import com.lssgoo.feature.auth.domain.model.Session

/**
 * Called once at app launch from the root screen. Two possible outcomes:
 *  - No tokens on disk → returns a Failure with UNAUTHORIZED, UI shows login.
 *  - Tokens present → call /auth/me, on success refresh the session cache.
 *    On failure the ApiClient's refresh path already kicked in once; if that
 *    failed too, tokens are cleared and UI falls back to login.
 *
 * Keeping this as a use case (vs inlined in the root ScreenModel) means tests
 * can stub it without touching network code.
 */
class BootstrapSessionUseCase(
    private val repository: AuthRepository,
) : NoParamUseCase<Session> {

    override suspend fun invoke(): AppResult<Session> {
        if (!repository.hasCredentials()) {
            return AppResult.Failure(
                com.lssgoo.core.result.AppError(
                    code = com.lssgoo.core.result.ErrorCode.UNAUTHORIZED,
                    message = "No active session.",
                ),
            )
        }
        return repository.fetchSession()
    }
}
