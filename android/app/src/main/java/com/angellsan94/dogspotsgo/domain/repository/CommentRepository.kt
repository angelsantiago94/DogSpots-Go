package com.angellsan94.dogspotsgo.domain.repository

import com.angellsan94.dogspotsgo.domain.model.Comment
import com.angellsan94.dogspotsgo.domain.model.RatingSummary

interface CommentRepository {
    suspend fun getComments(placeId: String): List<Comment>
    suspend fun getReplies(commentId: String): List<Comment>
    suspend fun createComment(placeId: String, body: String, replyToCommentId: String? = null): Comment
    suspend fun updateComment(commentId: String, body: String): Comment
    suspend fun deleteComment(commentId: String)
    suspend fun ratePlace(placeId: String, rating: Int): RatingSummary
    suspend fun deleteRating(placeId: String): RatingSummary
    suspend fun reportPlace(placeId: String, reason: String, details: String?)
    suspend fun reportComment(commentId: String, reason: String, details: String?)
}
