package com.angellsan94.dogspotsgo.data.remote

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import retrofit2.HttpException

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

inline fun <T> handleApiCall(json: Json = Json { ignoreUnknownKeys = true }, block: () -> T): T {
    return try {
        block()
    } catch (e: HttpException) {
        val errorBodyString = e.response()?.errorBody()?.string()
        val problemDetail = if (!errorBodyString.isNullOrBlank()) {
            try {
                json.decodeFromString<ProblemDetail>(errorBodyString)
            } catch (_: Exception) {
                null
            }
        } else {
            null
        }
        throw ApiException(
            statusCode = e.code(),
            errorBody = problemDetail,
            message = problemDetail?.detail ?: problemDetail?.title ?: when (e.code()) {
                401 -> "Credenciales incorrectas o sesión caducada."
                409 -> "El correo electrónico ya está registrado."
                422 -> "Los datos introducidos no son válidos."
                else -> "Error en la petición (Código ${e.code()})"
            }
        )
    }
}
