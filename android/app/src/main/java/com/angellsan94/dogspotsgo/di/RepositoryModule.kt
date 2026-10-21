package com.angellsan94.dogspotsgo.di

import com.angellsan94.dogspotsgo.data.repository.AuthRepositoryImpl
import com.angellsan94.dogspotsgo.data.repository.CommentRepositoryImpl
import com.angellsan94.dogspotsgo.data.repository.PlaceRepositoryImpl
import com.angellsan94.dogspotsgo.data.repository.UserRepositoryImpl
import com.angellsan94.dogspotsgo.domain.repository.AuthRepository
import com.angellsan94.dogspotsgo.domain.repository.CommentRepository
import com.angellsan94.dogspotsgo.domain.repository.PlaceRepository
import com.angellsan94.dogspotsgo.domain.repository.UserRepository
import dagger.Binds
import dagger.Module
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
abstract class RepositoryModule {

    @Binds
    @Singleton
    abstract fun bindAuthRepository(
        impl: AuthRepositoryImpl
    ): AuthRepository

    @Binds
    @Singleton
    abstract fun bindPlaceRepository(
        impl: PlaceRepositoryImpl
    ): PlaceRepository

    @Binds
    @Singleton
    abstract fun bindCommentRepository(
        impl: CommentRepositoryImpl
    ): CommentRepository

    @Binds
    @Singleton
    abstract fun bindUserRepository(
        impl: UserRepositoryImpl
    ): UserRepository
}
