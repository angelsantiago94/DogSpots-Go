package com.angellsan94.dogspotsgo.data.repository

import com.angellsan94.dogspotsgo.data.local.TokenDataSource
import com.angellsan94.dogspotsgo.data.remote.UsersApiService
import com.angellsan94.dogspotsgo.data.remote.handleApiCall
import com.angellsan94.dogspotsgo.data.remote.model.UpdateUserRequestDto
import com.angellsan94.dogspotsgo.domain.model.PublicUser
import com.angellsan94.dogspotsgo.domain.model.User
import com.angellsan94.dogspotsgo.domain.repository.UserRepository
import javax.inject.Inject

class UserRepositoryImpl @Inject constructor(
    private val usersApiService: UsersApiService,
    private val tokenDataSource: TokenDataSource
) : UserRepository {

    override suspend fun getMyProfile(): User = handleApiCall {
        val dto = usersApiService.getMyProfile()
        User(
            id = dto.id,
            email = dto.email,
            name = dto.name,
            createdAt = dto.created_at ?: ""
        )
    }

    override suspend fun updateMyProfile(name: String): User = handleApiCall {
        val dto = usersApiService.updateMyProfile(UpdateUserRequestDto(name = name))
        User(
            id = dto.id,
            email = dto.email,
            name = dto.name,
            createdAt = dto.created_at ?: ""
        )
    }

    override suspend fun deleteAccount(): Unit = handleApiCall {
        try {
            usersApiService.deleteAccount()
        } finally {
            tokenDataSource.clearTokens()
        }
    }

    override suspend fun getPublicProfile(userId: String): PublicUser = handleApiCall {
        val dto = usersApiService.getPublicProfile(userId)
        PublicUser(
            id = dto.id,
            name = dto.name
        )
    }
}
