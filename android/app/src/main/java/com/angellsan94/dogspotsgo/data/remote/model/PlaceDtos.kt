package com.angellsan94.dogspotsgo.data.remote.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
enum class PlaceCategoryDto {
    @SerialName("park") PARK,
    @SerialName("beach") BEACH,
    @SerialName("trail") TRAIL,
    @SerialName("dog_park") DOG_PARK,
    @SerialName("cafe") CAFE,
    @SerialName("restaurant") RESTAURANT,
    @SerialName("other") OTHER
}

@Serializable
data class PlaceSummaryDto(
    val id: String,
    val title: String,
    val lat: Double,
    val lng: Double,
    val category: PlaceCategoryDto,
    val cover_thumb_url: String? = null,
    val rating_avg: Double? = null,
    val rating_count: Int,
    val distance_m: Int? = null
)

@Serializable
data class PlaceListDto(
    val items: List<PlaceSummaryDto>,
    val next_cursor: String? = null
)
