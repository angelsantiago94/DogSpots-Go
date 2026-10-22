package com.angellsan94.dogspotsgo.data.remote

import com.angellsan94.dogspotsgo.data.local.TokenDataSource
import com.angellsan94.dogspotsgo.data.remote.model.RefreshRequestDto
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import okhttp3.Authenticator
import okhttp3.Request
import okhttp3.Response
import okhttp3.Route
import javax.inject.Inject

class TokenAuthenticator @Inject constructor(
    private val tokenDataSource: TokenDataSource,
    private val authApiLazy: dagger.Lazy<AuthApiService>
) : Authenticator {

    private val mutex = Any()

    override fun authenticate(route: Route?, response: Response): Request? {
        if (responseCount(response) >= 3) {
            return null
        }
        val failedToken = response.request
            .header("Authorization")
            ?.removePrefix("Bearer ")
            ?.trim()

        synchronized(mutex) {
            val currentAccessToken = runBlocking {
                tokenDataSource.accessTokenFlow.first()
            }
            if (!currentAccessToken.isNullOrBlank() && currentAccessToken != failedToken) {
                // Otro request ya renovó el token mientras esperábamos el lock.
                return response.request.newBuilder()
                    .header("Authorization", "Bearer $currentAccessToken")
                    .build()
            }

            val refreshToken = runBlocking {
                tokenDataSource.refreshTokenFlow.first()
            } ?: return null

            return try {
                val api = authApiLazy.get()
                val responseTokenPair = runBlocking {
                    api.refresh(RefreshRequestDto(refreshToken))
                }
                runBlocking {
                    tokenDataSource.saveTokens(responseTokenPair.access_token, responseTokenPair.refresh_token)
                }
                response.request.newBuilder()
                    .header("Authorization", "Bearer ${responseTokenPair.access_token}")
                    .build()
            } catch (e: Exception) {
                runBlocking {
                    tokenDataSource.clearTokens()
                }
                null
            }
        }
    }

    private fun responseCount(response: Response): Int {
        var count = 1
        var prior = response.priorResponse
        while (prior != null) {
            count++
            prior = prior.priorResponse
        }
        return count
    }
}
