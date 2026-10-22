package com.angellsan94.dogspotsgo.data.repository

import com.angellsan94.dogspotsgo.data.remote.CommentsApiService
import com.angellsan94.dogspotsgo.data.remote.model.CommentDto
import com.angellsan94.dogspotsgo.data.remote.model.CreateCommentRequestDto
import com.angellsan94.dogspotsgo.data.remote.model.CreateReportRequestDto
import com.angellsan94.dogspotsgo.data.remote.model.RatingRequestDto
import com.angellsan94.dogspotsgo.data.remote.model.UpdateCommentRequestDto
import com.angellsan94.dogspotsgo.domain.model.Comment
import com.angellsan94.dogspotsgo.domain.model.RatingSummary
import com.angellsan94.dogspotsgo.domain.repository.CommentRepository
import javax.inject.Inject

class CommentRepositoryImpl @Inject constructor(
    private val commentsApiService: CommentsApiService
) : CommentRepository {

    override suspend fun getComments(placeId: String): List<Comment> {
        // shortcut: cap at 10 pages, raise if a place can exceed 1000 comments
        val comments = mutableListOf<Comment>()
        var cursor: String? = null
        repeat(10) {
            val response = commentsApiService.getComments(placeId, limit = 100, cursor = cursor)
            comments += response.items.map { mapCommentDtoToDomain(it) }
            cursor = response.next_cursor
            if (cursor == null) return comments
        }
        return comments
    }

    override suspend fun getReplies(commentId: String): List<Comment> {
        // shortcut: cap at 10 pages, raise if a place can exceed 1000 comments
        val comments = mutableListOf<Comment>()
        var cursor: String? = null
        repeat(10) {
            val response = commentsApiService.getReplies(commentId, limit = 100, cursor = cursor)
            comments += response.items.map { mapCommentDtoToDomain(it) }
            cursor = response.next_cursor
            if (cursor == null) return comments
        }
        return comments
    }

    override suspend fun createComment(placeId: String, body: String, replyToCommentId: String?): Comment {
        val request = CreateCommentRequestDto(body = body, reply_to_comment_id = replyToCommentId)
        return mapCommentDtoToDomain(commentsApiService.createComment(placeId, request))
    }

    override suspend fun updateComment(commentId: String, body: String): Comment {
        val request = UpdateCommentRequestDto(body = body)
        return mapCommentDtoToDomain(commentsApiService.updateComment(commentId, request))
    }

    private fun mapCommentDtoToDomain(dto: CommentDto): Comment = Comment(
        id = dto.id,
        placeId = dto.place_id,
        userId = dto.author.id,
        authorName = dto.author.display_name,
        authorId = dto.author.id,
        parentId = dto.parent_id,
        replyToAuthorName = dto.reply_to?.display_name,
        body = dto.body,
        replyCount = dto.reply_count,
        edited = dto.edited,
        deleted = dto.deleted,
        createdAt = dto.created_at
    )

    override suspend fun deleteComment(commentId: String) {
        commentsApiService.deleteComment(commentId)
    }

    override suspend fun ratePlace(placeId: String, rating: Int): RatingSummary {
        val request = RatingRequestDto(rating = rating)
        val dto = commentsApiService.ratePlace(placeId, request)
        return RatingSummary(
            ratingAvg = dto.rating_avg,
            ratingCount = dto.rating_count,
            myRating = dto.my_rating
        )
    }

    override suspend fun deleteRating(placeId: String): RatingSummary {
        val dto = commentsApiService.deleteRating(placeId)
        return RatingSummary(
            ratingAvg = dto.rating_avg,
            ratingCount = dto.rating_count,
            myRating = dto.my_rating
        )
    }

    override suspend fun reportPlace(placeId: String, reason: String, details: String?) {
        val request = CreateReportRequestDto(reason = reason, details = details)
        commentsApiService.reportPlace(placeId, request)
    }

    override suspend fun reportComment(commentId: String, reason: String, details: String?) {
        val request = CreateReportRequestDto(reason = reason, details = details)
        commentsApiService.reportComment(commentId, request)
    }
}
