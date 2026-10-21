package com.angellsan94.dogspotsgo.data.remote.model

import kotlinx.serialization.Serializable

@Serializable
data class UserDto(
    val id: String,
    val email: String,
    val name: String,
    val created_at: String
)

@Serializable
data class TokenPairDto(
    val access_token: String,
    val refresh_token: String
)

@Serializable
data class AuthResponseDto(
    val user: UserDto,
    val tokens: TokenPairDto
)

@Serializable
data class LoginRequestDto(
    val email: String,
    val password: String
)

@Serializable
data class RegisterRequestDto(
    val email: String,
    val password: String,
    val name: String
)

@Serializable
data class RefreshRequestDto(
    val refresh_token: String
)
