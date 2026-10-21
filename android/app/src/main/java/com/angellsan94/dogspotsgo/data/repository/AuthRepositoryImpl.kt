package com.angellsan94.dogspotsgo.data.repository

import com.angellsan94.dogspotsgo.data.local.TokenDataSource
import com.angellsan94.dogspotsgo.data.remote.AuthApiService
import com.angellsan94.dogspotsgo.data.remote.UsersApiService
import com.angellsan94.dogspotsgo.data.remote.handleApiCall
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
    private val usersApiService: UsersApiService,
    private val tokenDataSource: TokenDataSource
) : AuthRepository {

    override val accessTokenFlow: Flow<String?> = tokenDataSource.accessTokenFlow
    override val refreshTokenFlow: Flow<String?> = tokenDataSource.refreshTokenFlow

    override suspend fun login(email: String, password: String): User = handleApiCall {
        val response = authApiService.login(LoginRequestDto(email, password))
        tokenDataSource.saveTokens(response.access_token, response.refresh_token)
        try {
            val profile = usersApiService.getMyProfile()
            User(
                id = profile.id,
                email = profile.email,
                name = profile.name,
                createdAt = profile.created_at ?: ""
            )
        } catch (_: Exception) {
            User(
                id = response.user_id,
                email = email,
                name = "",
                createdAt = ""
            )
        }
    }

    override suspend fun register(email: String, password: String, name: String): User = handleApiCall {
        val response = authApiService.register(RegisterRequestDto(email, password, name))
        tokenDataSource.saveTokens(response.access_token, response.refresh_token)
        User(
            id = response.user_id,
            email = email,
            name = name,
            createdAt = ""
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
