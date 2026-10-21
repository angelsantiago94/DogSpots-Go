package com.angellsan94.dogspotsgo.data.remote.model

import kotlinx.serialization.Serializable

@Serializable
data class TagDto(
    val slug: String,
    val label: String
)

@Serializable
data class TagListDto(
    val items: List<TagDto>
)

@Serializable
data class PhotoDto(
    val id: String,
    val url: String,
    val thumb_url: String,
    val caption: String? = null,
    val width: Int,
    val height: Int,
    val position: Int,
    val uploaded_by: String,
    val created_at: String
)

@Serializable
data class PublicUserDto(
    val id: String,
    val name: String
)

@Serializable
data class PlaceAuthorDto(
    val id: String,
    val display_name: String
)

@Serializable
data class PlaceDetailDto(
    val id: String,
    val title: String,
    val description: String? = null,
    val lat: Double,
    val lng: Double,
    val category: PlaceCategoryDto,
    val tags: List<TagDto> = emptyList(),
    val photos: List<PhotoDto> = emptyList(),
    val rating_avg: Double? = null,
    val rating_count: Int,
    val my_rating: Int? = null,
    val comment_count: Int,
    val distance_m: Int? = null,
    val author: PlaceAuthorDto,
    val created_at: String,
    val updated_at: String
)

@Serializable
data class CreatePlaceRequestDto(
    val title: String,
    val description: String? = null,
    val lat: Double,
    val lng: Double,
    val category: PlaceCategoryDto,
    val tags: List<String> = emptyList()
)

@Serializable
data class UpdatePlaceRequestDto(
    val title: String? = null,
    val description: String? = null,
    val lat: Double? = null,
    val lng: Double? = null,
    val category: PlaceCategoryDto? = null,
    val tags: List<String>? = null
)
