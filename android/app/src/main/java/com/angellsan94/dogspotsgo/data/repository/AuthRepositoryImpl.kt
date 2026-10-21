package com.angellsan94.dogspotsgo.data.repository

import com.angellsan94.dogspotsgo.data.local.TokenDataSource
import com.angellsan94.dogspotsgo.data.remote.AuthApiService
import com.angellsan94.dogspotsgo.data.remote.model.LoginRequestDto
import com.angellsan94.dogspotsgo.data.remote.model.RefreshRequestDto
import com.angellsan94.dogspotsgo.data.remote.model.RegisterRequestDto
import com.angellsan94.dogspotsgo.domain.model.User
import com.angellsan94.dogspotsgo.domain.repository.AuthRepository
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import javax.inject.Inject

class AuthRepositoryImpl @Inject constructor(
    private val authApiService: AuthApiService,
    private val tokenDataSource: TokenDataSource
) : AuthRepository {

    override val accessTokenFlow: Flow<String?> = tokenDataSource.accessTokenFlow
    override val refreshTokenFlow: Flow<String?> = tokenDataSource.refreshTokenFlow

    override suspend fun login(email: String, password: String): User {
        val response = authApiService.login(LoginRequestDto(email, password))
        tokenDataSource.saveTokens(response.tokens.access_token, response.tokens.refresh_token)
        return User(
            id = response.user.id,
            email = response.user.email,
            name = response.user.name,
            createdAt = response.user.created_at
        )
    }

    override suspend fun register(email: String, password: String, name: String): User {
        val response = authApiService.register(RegisterRequestDto(email, password, name))
        tokenDataSource.saveTokens(response.tokens.access_token, response.tokens.refresh_token)
        return User(
            id = response.user.id,
            email = response.user.email,
            name = response.user.name,
            createdAt = response.user.created_at
        )
    }

    override suspend fun logout() {
        try {
            val refreshToken = tokenDataSource.refreshTokenFlow.first()
            if (!refreshToken.isNullOrBlank()) {
                authApiService.logout(RefreshRequestDto(refreshToken))
            }
        } catch (_: Exception) {
            // Ignore network errors on logout
        } finally {
            tokenDataSource.clearTokens()
        }
    }
}
