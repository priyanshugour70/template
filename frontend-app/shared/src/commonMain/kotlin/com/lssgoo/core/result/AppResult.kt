package com.lssgoo.core.result

/**
 * Sealed result type. We deliberately avoid [kotlin.Result] because:
 *  - It boxes exceptions (we want a structured [AppError] instead).
 *  - It's a value class with awkward generics on KMP / iOS.
 *
 * Pattern: every repository / use case returns `AppResult<T>`. Presentation
 * layer pattern-matches with `when (result) { is Success -> ... ; is Failure -> ... }`.
 */
sealed class AppResult<out T> {
    data class Success<T>(val data: T) : AppResult<T>()
    data class Failure(val error: AppError) : AppResult<Nothing>()

    inline fun <R> map(transform: (T) -> R): AppResult<R> = when (this) {
        is Success -> Success(transform(data))
        is Failure -> this
    }

    inline fun <R> flatMap(transform: (T) -> AppResult<R>): AppResult<R> = when (this) {
        is Success -> transform(data)
        is Failure -> this
    }

    inline fun onSuccess(block: (T) -> Unit): AppResult<T> {
        if (this is Success) block(data)
        return this
    }

    inline fun onFailure(block: (AppError) -> Unit): AppResult<T> {
        if (this is Failure) block(error)
        return this
    }

    fun getOrNull(): T? = (this as? Success)?.data
    fun errorOrNull(): AppError? = (this as? Failure)?.error
}
