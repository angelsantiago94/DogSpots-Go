package com.angellsan94.dogspotsgo.di

import com.angellsan94.dogspotsgo.BuildConfig
import com.angellsan94.dogspotsgo.data.local.TokenDataSource
import com.angellsan94.dogspotsgo.data.remote.AuthApiService
import com.angellsan94.dogspotsgo.data.remote.AuthInterceptor
import com.angellsan94.dogspotsgo.data.remote.CommentsApiService
import com.angellsan94.dogspotsgo.data.remote.PlacesApiService
import com.angellsan94.dogspotsgo.data.remote.TokenAuthenticator
import com.angellsan94.dogspotsgo.data.remote.UsersApiService
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.kotlinx.serialization.asConverterFactory
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object NetworkModule {

    @Provides
    @Singleton
    fun provideJson(): Json = Json {
        ignoreUnknownKeys = true
        coerceInputValues = true
        isLenient = true
    }

    @Provides
    @Singleton
    fun provideAuthInterceptor(tokenDataSource: TokenDataSource): AuthInterceptor {
        return AuthInterceptor(tokenDataSource)
    }

    @Provides
    @Singleton
    fun provideTokenAuthenticator(
        tokenDataSource: TokenDataSource,
        authApiLazy: dagger.Lazy<AuthApiService>
    ): TokenAuthenticator {
        return TokenAuthenticator(tokenDataSource, authApiLazy)
    }

    @Provides
    @Singleton
    fun provideOkHttpClient(
        authInterceptor: AuthInterceptor,
        tokenAuthenticator: TokenAuthenticator
    ): OkHttpClient {
        val logging = HttpLoggingInterceptor().apply {
            level = HttpLoggingInterceptor.Level.BODY
        }
        return OkHttpClient.Builder()
            .addInterceptor(authInterceptor)
            .authenticator(tokenAuthenticator)
            .addInterceptor(logging)
            .build()
    }

    @Provides
    @Singleton
    fun provideRetrofit(okHttpClient: OkHttpClient, json: Json): Retrofit {
        val contentType = "application/json".toMediaType()
        return Retrofit.Builder()
            .baseUrl(BuildConfig.BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(json.asConverterFactory(contentType))
            .build()
    }

    @Provides
    @Singleton
    fun provideAuthApiService(retrofit: Retrofit): AuthApiService {
        return retrofit.create(AuthApiService::class.java)
    }

    @Provides
    @Singleton
    fun providePlacesApiService(retrofit: Retrofit): PlacesApiService {
        return retrofit.create(PlacesApiService::class.java)
    }

    @Provides
    @Singleton
    fun provideCommentsApiService(retrofit: Retrofit): CommentsApiService {
        return retrofit.create(CommentsApiService::class.java)
    }

    @Provides
    @Singleton
    fun provideUsersApiService(retrofit: Retrofit): UsersApiService {
        return retrofit.create(UsersApiService::class.java)
    }
}
