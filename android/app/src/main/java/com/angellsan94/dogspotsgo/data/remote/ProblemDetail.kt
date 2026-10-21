package com.angellsan94.dogspotsgo.data.remote

import kotlinx.serialization.Serializable

@Serializable
data class ProblemDetail(
    val type: String? = null,
    val title: String? = null,
    val status: Int? = null,
    val detail: String? = null,
    val instance: String? = null,
    val code: String? = null
)

class ApiException(
    val statusCode: Int,
    val errorBody: ProblemDetail?,
    message: String?
) : Exception(message ?: errorBody?.detail ?: errorBody?.title ?: "Error HTTP $statusCode")
