package com.angellsan94.dogspotsgo.domain.repository

import com.angellsan94.dogspotsgo.domain.model.PublicUser
import com.angellsan94.dogspotsgo.domain.model.User

interface UserRepository {
    suspend fun getMyProfile(): User
    suspend fun updateMyProfile(name: String): User
    suspend fun deleteAccount()
    suspend fun getPublicProfile(userId: String): PublicUser
}
