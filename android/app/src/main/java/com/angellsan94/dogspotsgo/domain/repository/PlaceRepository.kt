package com.angellsan94.dogspotsgo.domain.repository

import com.angellsan94.dogspotsgo.domain.model.Place
import com.angellsan94.dogspotsgo.domain.model.PlaceDetail
import com.angellsan94.dogspotsgo.domain.model.Tag

interface PlaceRepository {
    suspend fun getPlacesByBbox(
        minLat: Double,
        minLng: Double,
        maxLat: Double,
        maxLng: Double,
        category: String? = null,
        minRating: Double? = null
    ): List<Place>

    suspend fun getPlaceDetail(placeId: String): PlaceDetail

    suspend fun createPlace(
        title: String,
        description: String?,
        lat: Double,
        lng: Double,
        category: String,
        tags: List<String>
    ): PlaceDetail

    suspend fun updatePlace(
        placeId: String,
        title: String?,
        description: String?,
        lat: Double?,
        lng: Double?,
        category: String?,
        tags: List<String>?
    ): PlaceDetail

    suspend fun deletePlace(placeId: String)

    suspend fun getMyPlaces(): List<Place>

    suspend fun getTags(): List<Tag>
}
