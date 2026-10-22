package com.angellsan94.dogspotsgo.data.remote

import com.angellsan94.dogspotsgo.data.remote.model.*
import okhttp3.MultipartBody
import okhttp3.RequestBody
import retrofit2.Response
import retrofit2.http.*

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

    @GET("places/{placeId}")
    suspend fun getPlaceDetail(@Path("placeId") placeId: String): PlaceDetailDto

    @POST("places")
    suspend fun createPlace(@Body request: CreatePlaceRequestDto): PlaceDetailDto

    @PATCH("places/{placeId}")
    suspend fun updatePlace(
        @Path("placeId") placeId: String,
        @Body request: UpdatePlaceRequestDto
    ): PlaceDetailDto

    @DELETE("places/{placeId}")
    suspend fun deletePlace(@Path("placeId") placeId: String): Response<Unit>

    @GET("users/me/places")
    suspend fun getMyPlaces(
        @Query("limit") limit: Int? = null,
        @Query("cursor") cursor: String? = null
    ): PlaceListDto

    @GET("tags")
    suspend fun getTags(): TagListDto

    @Multipart
    @POST("places/{placeId}/photos")
    suspend fun uploadPhoto(
        @Path("placeId") placeId: String,
        @Part file: MultipartBody.Part,
        @Part("caption") caption: RequestBody? = null
    ): PhotoDto

    @DELETE("places/{placeId}/photos/{photoId}")
    suspend fun deletePhoto(
        @Path("placeId") placeId: String,
        @Path("photoId") photoId: String
    )
}
