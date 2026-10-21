package com.angellsan94.dogspotsgo.domain.repository

import com.angellsan94.dogspotsgo.domain.model.User
import kotlinx.coroutines.flow.Flow

interface AuthRepository {
    val accessTokenFlow: Flow<String?>
    val refreshTokenFlow: Flow<String?>

    suspend fun login(email: String, password: String): User
    suspend fun register(email: String, password: String, name: String): User
    suspend fun logout()
}
