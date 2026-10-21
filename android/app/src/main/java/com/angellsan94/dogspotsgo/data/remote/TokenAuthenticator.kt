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

        val refreshToken = runBlocking {
            tokenDataSource.refreshTokenFlow.first()
        } ?: return null

        synchronized(mutex) {
            try {
                val api = authApiLazy.get()
                val responseTokenPair = runBlocking {
                    api.refresh(RefreshRequestDto(refreshToken))
                }
                runBlocking {
                    tokenDataSource.saveTokens(responseTokenPair.access_token, responseTokenPair.refresh_token)
                }
                return response.request.newBuilder()
                    .header("Authorization", "Bearer ${responseTokenPair.access_token}")
                    .build()
            } catch (e: Exception) {
                runBlocking {
                    tokenDataSource.clearTokens()
                }
                return null
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
