package com.angellsan94.dogspotsgo.data.remote.model

import kotlinx.serialization.Serializable

@Serializable
data class CommentDto(
    val id: String,
    val place_id: String,
    val user_id: String,
    val author: PlaceAuthorDto,
    val parent_id: String? = null,
    val reply_to: PlaceAuthorDto? = null,
    val body: String? = null,
    val reply_count: Int = 0,
    val edited: Boolean = false,
    val deleted: Boolean = false,
    val created_at: String,
    val updated_at: String
)

@Serializable
data class CommentListDto(
    val items: List<CommentDto>,
    val next_cursor: String? = null
)

@Serializable
data class CreateCommentRequestDto(
    val body: String,
    val reply_to_comment_id: String? = null
)

@Serializable
data class UpdateCommentRequestDto(
    val body: String
)

@Serializable
data class RatingRequestDto(
    val rating: Int
)

@Serializable
data class RatingSummaryDto(
    val rating_avg: Double? = null,
    val rating_count: Int,
    val my_rating: Int? = null
)

@Serializable
data class CreateReportRequestDto(
    val reason: String,
    val details: String? = null
)
