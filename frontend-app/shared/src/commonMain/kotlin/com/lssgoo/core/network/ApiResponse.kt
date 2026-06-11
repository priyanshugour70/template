package com.lssgoo.core.network

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * Wire-format envelope produced by the Go backend's `internal/pkg/response`.
 *
 * Every endpoint returns:
 * ```
 * { "success": bool,
 *   "data":    T?,           // payload on success
 *   "error":   ErrPayload?,  // structured error on failure
 *   "message": string?,
 *   "timestamp": string?,
 *   "pagination": Pagination? }
 * ```
 *
 * We model it generically; the [ApiClient] unwraps it before any feature code sees it.
 */
@Serializable
data class ApiResponse<T>(
    val success: Boolean = false,
    val data: T? = null,
    val error: ApiErrorPayload? = null,
    val message: String? = null,
    val timestamp: String? = null,
    val pagination: ApiPagination? = null,
)

@Serializable
data class ApiErrorPayload(
    val code: String? = null,
    val message: String? = null,
    val details: Map<String, String>? = null,
)

@Serializable
data class ApiPagination(
    val page: Int = 1,
    val limit: Int = 20,
    val total: Long = 0,
    @SerialName("totalPages") val totalPages: Int = 0,
    @SerialName("hasNext") val hasNext: Boolean = false,
    @SerialName("hasPrev") val hasPrev: Boolean = false,
)
