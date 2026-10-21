package com.angellsan94.dogspotsgo.domain.model

data class Comment(
    val id: String,
    val placeId: String,
    val userId: String,
    val authorName: String,
    val authorId: String,
    val parentId: String?,
    val replyToAuthorName: String?,
    val body: String?,
    val replyCount: Int,
    val edited: Boolean,
    val deleted: Boolean,
    val createdAt: String,
    val replies: List<Comment> = emptyList()
)

data class RatingSummary(
    val ratingAvg: Double?,
    val ratingCount: Int,
    val myRating: Int?
)
