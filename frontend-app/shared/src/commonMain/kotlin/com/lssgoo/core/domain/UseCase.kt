package com.lssgoo.core.domain

import com.lssgoo.core.result.AppResult

/**
 * Marker for a single-purpose application action ("verb"). Following the clean-arch
 * convention from the Go backend's service layer:
 *  - Stateless.
 *  - Inputs in, [AppResult] out.
 *  - Holds references to repositories / other use cases via constructor.
 *
 * Use cases are wired in their feature's Koin module.
 *
 * Subclassing is optional — interfaces work too. The marker exists so feature
 * directories all look the same: `feature/x/domain/usecase/DoSomethingUseCase.kt`.
 */
interface UseCase<in P, out T> {
    suspend operator fun invoke(params: P): AppResult<T>
}

/** Convenience for use cases that take no parameters. */
interface NoParamUseCase<out T> {
    suspend operator fun invoke(): AppResult<T>
}
