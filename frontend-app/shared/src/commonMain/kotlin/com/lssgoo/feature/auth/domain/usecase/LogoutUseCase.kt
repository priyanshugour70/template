package com.lssgoo.feature.auth.domain.usecase

import com.lssgoo.core.domain.NoParamUseCase
import com.lssgoo.core.result.AppResult
import com.lssgoo.feature.auth.domain.AuthRepository

class LogoutUseCase(private val repository: AuthRepository) : NoParamUseCase<Unit> {
    override suspend fun invoke(): AppResult<Unit> = repository.logout()
}
