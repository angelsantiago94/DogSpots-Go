package com.angellsan94.dogspotsgo.data.remote

import com.angellsan94.dogspotsgo.data.remote.model.PlaceListDto
import retrofit2.http.GET
import retrofit2.http.Query

interface PlacesApiService {
    @GET("places")
    suspend fun getPlaces(
        @Query("min_lat") minLat: Double? = null,
        @Query("min_lng") minLng: Double? = null,
        @Query("max_lat") maxLat: Double? = null,
        @Query("max_lng") maxLng: Double? = null,
        @Query("lat") lat: Double? = null,
        @Query("lng") lng: Double? = null,
        @Query("radius_m") radiusM: Int? = null,
        @Query("category") category: String? = null,
        @Query("tags") tags: String? = null,
        @Query("min_rating") minRating: Double? = null,
        @Query("q") q: String? = null,
        @Query("sort") sort: String? = null,
        @Query("limit") limit: Int? = null,
        @Query("cursor") cursor: String? = null
    ): PlaceListDto
}
