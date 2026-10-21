package com.angellsan94.dogspotsgo.data.remote

import com.angellsan94.dogspotsgo.data.remote.model.*
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.POST

interface AuthApiService {
    @POST("auth/register")
    suspend fun register(@Body request: RegisterRequestDto): AuthResponseDto

    @POST("auth/login")
    suspend fun login(@Body request: LoginRequestDto): AuthResponseDto

    @POST("auth/refresh")
    suspend fun refresh(@Body request: RefreshRequestDto): TokenPairDto

    @POST("auth/logout")
    suspend fun logout(@Body request: RefreshRequestDto): Response<Unit>
}
