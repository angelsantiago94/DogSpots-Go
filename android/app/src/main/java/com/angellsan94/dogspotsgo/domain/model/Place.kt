package com.angellsan94.dogspotsgo.domain.model

enum class PlaceCategory {
    PARK, BEACH, TRAIL, DOG_PARK, CAFE, RESTAURANT, OTHER
}

data class Place(
    val id: String,
    val title: String,
    val lat: Double,
    val lng: Double,
    val category: PlaceCategory,
    val coverThumbUrl: String?,
    val ratingAvg: Double?,
    val ratingCount: Int,
    val distanceM: Int?
)
