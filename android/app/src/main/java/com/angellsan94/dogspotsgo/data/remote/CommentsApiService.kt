package com.angellsan94.dogspotsgo.data.remote

import com.angellsan94.dogspotsgo.data.remote.model.*
import retrofit2.Response
import retrofit2.http.*

interface CommentsApiService {
    @GET("places/{placeId}/comments")
    suspend fun getComments(
        @Path("placeId") placeId: String,
        @Query("sort") sort: String? = null,
        @Query("limit") limit: Int? = null,
        @Query("cursor") cursor: String? = null
    ): CommentListDto

    @POST("places/{placeId}/comments")
    suspend fun createComment(
        @Path("placeId") placeId: String,
        @Body request: CreateCommentRequestDto
    ): CommentDto

    @GET("comments/{commentId}/replies")
    suspend fun getReplies(
        @Path("commentId") commentId: String,
        @Query("limit") limit: Int? = null,
        @Query("cursor") cursor: String? = null
    ): CommentListDto

    @PATCH("comments/{commentId}")
    suspend fun updateComment(
        @Path("commentId") commentId: String,
        @Body request: UpdateCommentRequestDto
    ): CommentDto

    @DELETE("comments/{commentId}")
    suspend fun deleteComment(@Path("commentId") commentId: String): Response<Unit>

    @PUT("places/{placeId}/rating")
    suspend fun ratePlace(
        @Path("placeId") placeId: String,
        @Body request: RatingRequestDto
    ): RatingSummaryDto

    @DELETE("places/{placeId}/rating")
    suspend fun deleteRating(@Path("placeId") placeId: String): RatingSummaryDto

    @POST("places/{placeId}/reports")
    suspend fun reportPlace(
        @Path("placeId") placeId: String,
        @Body request: CreateReportRequestDto
    ): Response<Unit>

    @POST("comments/{commentId}/reports")
    suspend fun reportComment(
        @Path("commentId") commentId: String,
        @Body request: CreateReportRequestDto
    ): Response<Unit>
}
