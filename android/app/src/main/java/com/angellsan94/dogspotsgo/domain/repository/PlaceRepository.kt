package com.angellsan94.dogspotsgo.domain.repository

import com.angellsan94.dogspotsgo.domain.model.Place

interface PlaceRepository {
    suspend fun getPlacesByBbox(
        minLat: Double,
        minLng: Double,
        maxLat: Double,
        maxLng: Double,
        category: String? = null,
        minRating: Double? = null
    ): List<Place>
}
