package com.lssgoo.core.result

/**
 * Domain-level error mirroring the Go backend's `apperr.AppError`.
 *
 * The backend emits errors as `{ code, message, details }` inside the envelope's
 * `error` field. We preserve that structure end-to-end so a UI layer can choose
 * to show [message] verbatim or branch on a stable [code] (e.g. INVALID_CREDENTIALS).
 */
data class AppError(
    val code: ErrorCode,
    val message: String,
    val details: Map<String, String> = emptyMap(),
    val cause: Throwable? = null,
)

/**
 * Stable error codes that flow from backend → UI. New codes can be added without
 * affecting older clients because [UNKNOWN] is the safe fallback for unrecognised
 * server strings.
 */
enum class ErrorCode(val raw: String) {
    NETWORK("NETWORK"),
    TIMEOUT("TIMEOUT"),
    UNAUTHORIZED("UNAUTHORIZED"),
    FORBIDDEN("FORBIDDEN"),
    INVALID_CREDENTIALS("INVALID_CREDENTIALS"),
    TOKEN_EXPIRED("TOKEN_EXPIRED"),
    INVALID_TOKEN("INVALID_TOKEN"),
    NOT_FOUND("NOT_FOUND"),
    CONFLICT("CONFLICT"),
    VALIDATION("VALIDATION"),
    RATE_LIMITED("RATE_LIMITED"),
    INTERNAL("INTERNAL"),
    UNKNOWN("UNKNOWN");

    companion object {
        fun fromRaw(raw: String?): ErrorCode = entries.firstOrNull { it.raw == raw } ?: UNKNOWN
    }
}
