package com.angellsan94.dogspotsgo.domain.model

data class PlaceDetail(
    val id: String,
    val title: String,
    val description: String?,
    val lat: Double,
    val lng: Double,
    val category: PlaceCategory,
    val tags: List<Tag>,
    val photos: List<Photo>,
    val ratingAvg: Double?,
    val ratingCount: Int,
    val myRating: Int?,
    val commentCount: Int,
    val authorName: String,
    val authorId: String,
    val createdAt: String
)
