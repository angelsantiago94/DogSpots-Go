package com.angellsan94.dogspotsgo.data.remote

import com.angellsan94.dogspotsgo.data.remote.model.PublicUserDto
import com.angellsan94.dogspotsgo.data.remote.model.UpdateUserRequestDto
import com.angellsan94.dogspotsgo.data.remote.model.UserDto
import retrofit2.Response
import retrofit2.http.*

interface UsersApiService {
    @GET("users/me")
    suspend fun getMyProfile(): UserDto

    @PATCH("users/me")
    suspend fun updateMyProfile(@Body request: UpdateUserRequestDto): UserDto

    @DELETE("users/me")
    suspend fun deleteAccount(): Response<Unit>

    @GET("users/{userId}")
    suspend fun getPublicProfile(@Path("userId") userId: String): PublicUserDto
}
