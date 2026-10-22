package com.angellsan94.dogspotsgo.data.repository

import com.angellsan94.dogspotsgo.BuildConfig
import com.angellsan94.dogspotsgo.data.remote.PlacesApiService
import com.angellsan94.dogspotsgo.data.remote.handleApiCall
import com.angellsan94.dogspotsgo.data.remote.model.CreatePlaceRequestDto
import com.angellsan94.dogspotsgo.data.remote.model.PlaceCategoryDto
import com.angellsan94.dogspotsgo.data.remote.model.UpdatePlaceRequestDto
import com.angellsan94.dogspotsgo.domain.model.*
import com.angellsan94.dogspotsgo.domain.repository.PlaceRepository
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.MultipartBody
import okhttp3.RequestBody.Companion.toRequestBody
import javax.inject.Inject

class PlaceRepositoryImpl @Inject constructor(
    private val placesApiService: PlacesApiService
) : PlaceRepository {

    override suspend fun getPlacesByBbox(
        minLat: Double,
        minLng: Double,
        maxLat: Double,
        maxLng: Double,
        category: String?,
        minRating: Double?
    ): List<Place> = handleApiCall {
        val response = placesApiService.getPlaces(
            minLat = minLat,
            minLng = minLng,
            maxLat = maxLat,
            maxLng = maxLng,
            category = category,
            minRating = minRating
        )
        response.items.map { dto ->
            Place(
                id = dto.id,
                title = dto.title,
                lat = dto.lat,
                lng = dto.lng,
                category = mapCategoryDtoToDomain(dto.category),
                coverThumbUrl = dto.cover_thumb_url.toFullMediaUrl(),
                ratingAvg = dto.rating_avg,
                ratingCount = dto.rating_count,
                distanceM = dto.distance_m
            )
        }
    }

    override suspend fun getPlaceDetail(placeId: String): PlaceDetail = handleApiCall {
        val dto = placesApiService.getPlaceDetail(placeId)
        PlaceDetail(
            id = dto.id,
            title = dto.title,
            description = dto.description,
            lat = dto.lat,
            lng = dto.lng,
            category = mapCategoryDtoToDomain(dto.category),
            tags = dto.tags.map { Tag(it.slug, it.label) },
            photos = dto.photos.map {
                Photo(
                    id = it.id,
                    url = it.url.toFullMediaUrl() ?: "",
                    thumbUrl = it.thumb_url.toFullMediaUrl() ?: "",
                    caption = it.caption,
                    width = it.width,
                    height = it.height,
                    position = it.position,
                    uploadedBy = it.uploaded_by,
                    createdAt = it.created_at
                )
            },
            ratingAvg = dto.rating_avg,
            ratingCount = dto.rating_count,
            myRating = dto.my_rating,
            commentCount = dto.comment_count,
            authorName = dto.author.display_name,
            authorId = dto.author.id,
            createdAt = dto.created_at
        )
    }

    override suspend fun createPlace(
        title: String,
        description: String?,
        lat: Double,
        lng: Double,
        category: String,
        tags: List<String>
    ): PlaceDetail = handleApiCall {
        val categoryDto = mapCategoryDomainToDto(category)
        val request = CreatePlaceRequestDto(
            title = title,
            description = description,
            lat = lat,
            lng = lng,
            category = categoryDto,
            tags = tags
        )
        val dto = placesApiService.createPlace(request)
        PlaceDetail(
            id = dto.id,
            title = dto.title,
            description = dto.description,
            lat = dto.lat,
            lng = dto.lng,
            category = mapCategoryDtoToDomain(dto.category),
            tags = dto.tags.map { Tag(it.slug, it.label) },
            photos = dto.photos.map {
                Photo(
                    id = it.id,
                    url = it.url.toFullMediaUrl() ?: "",
                    thumbUrl = it.thumb_url.toFullMediaUrl() ?: "",
                    caption = it.caption,
                    width = it.width,
                    height = it.height,
                    position = it.position,
                    uploadedBy = it.uploaded_by,
                    createdAt = it.created_at
                )
            },
            ratingAvg = dto.rating_avg,
            ratingCount = dto.rating_count,
            myRating = dto.my_rating,
            commentCount = dto.comment_count,
            authorName = dto.author.display_name,
            authorId = dto.author.id,
            createdAt = dto.created_at
        )
    }

    override suspend fun updatePlace(
        placeId: String,
        title: String?,
        description: String?,
        lat: Double?,
        lng: Double?,
        category: String?,
        tags: List<String>?
    ): PlaceDetail = handleApiCall {
        val categoryDto = category?.let { mapCategoryDomainToDto(it) }
        val request = UpdatePlaceRequestDto(
            title = title,
            description = description,
            lat = lat,
            lng = lng,
            category = categoryDto,
            tags = tags
        )
        val dto = placesApiService.updatePlace(placeId, request)
        PlaceDetail(
            id = dto.id,
            title = dto.title,
            description = dto.description,
            lat = dto.lat,
            lng = dto.lng,
            category = mapCategoryDtoToDomain(dto.category),
            tags = dto.tags.map { Tag(it.slug, it.label) },
            photos = dto.photos.map {
                Photo(
                    id = it.id,
                    url = it.url.toFullMediaUrl() ?: "",
                    thumbUrl = it.thumb_url.toFullMediaUrl() ?: "",
                    caption = it.caption,
                    width = it.width,
                    height = it.height,
                    position = it.position,
                    uploadedBy = it.uploaded_by,
                    createdAt = it.created_at
                )
            },
            ratingAvg = dto.rating_avg,
            ratingCount = dto.rating_count,
            myRating = dto.my_rating,
            commentCount = dto.comment_count,
            authorName = dto.author.display_name,
            authorId = dto.author.id,
            createdAt = dto.created_at
        )
    }

    override suspend fun uploadPhoto(
        placeId: String,
        image: ByteArray,
        fileName: String,
        mimeType: String
    ): Photo = handleApiCall {
        val filePart = MultipartBody.Part.createFormData(
            "file",
            fileName,
            image.toRequestBody(mimeType.toMediaTypeOrNull())
        )
        val dto = placesApiService.uploadPhoto(placeId, filePart)
        Photo(
            id = dto.id,
            url = dto.url.toFullMediaUrl() ?: "",
            thumbUrl = dto.thumb_url.toFullMediaUrl() ?: "",
            caption = dto.caption,
            width = dto.width,
            height = dto.height,
            position = dto.position,
            uploadedBy = dto.uploaded_by,
            createdAt = dto.created_at
        )
    }

    override suspend fun deletePlace(placeId: String) = handleApiCall {
        placesApiService.deletePlace(placeId)
        Unit
    }

    override suspend fun deletePhoto(placeId: String, photoId: String) = handleApiCall {
        placesApiService.deletePhoto(placeId, photoId)
        Unit
    }

    override suspend fun getMyPlaces(): List<Place> = handleApiCall {
        val response = placesApiService.getMyPlaces()
        response.items.map { dto ->
            Place(
                id = dto.id,
                title = dto.title,
                lat = dto.lat,
                lng = dto.lng,
                category = mapCategoryDtoToDomain(dto.category),
                coverThumbUrl = dto.cover_thumb_url.toFullMediaUrl(),
                ratingAvg = dto.rating_avg,
                ratingCount = dto.rating_count,
                distanceM = dto.distance_m
            )
        }
    }

    override suspend fun getTags(): List<Tag> = handleApiCall {
        val response = placesApiService.getTags()
        response.items.map { Tag(it.slug, it.label) }
    }

    private fun mapCategoryDtoToDomain(dto: PlaceCategoryDto): PlaceCategory {
        return when (dto) {
            PlaceCategoryDto.PARK -> PlaceCategory.PARK
            PlaceCategoryDto.BEACH -> PlaceCategory.BEACH
            PlaceCategoryDto.TRAIL -> PlaceCategory.TRAIL
            PlaceCategoryDto.DOG_PARK -> PlaceCategory.DOG_PARK
            PlaceCategoryDto.CAFE -> PlaceCategory.CAFE
            PlaceCategoryDto.RESTAURANT -> PlaceCategory.RESTAURANT
            PlaceCategoryDto.OTHER -> PlaceCategory.OTHER
        }
    }

    private fun mapCategoryDomainToDto(category: String): PlaceCategoryDto {
        return when (category.lowercase()) {
            "park", "parque" -> PlaceCategoryDto.PARK
            "beach", "playa" -> PlaceCategoryDto.BEACH
            "trail", "ruta_natural" -> PlaceCategoryDto.TRAIL
            "dog_park", "pipican_area_canina", "pipican", "area_canina" -> PlaceCategoryDto.DOG_PARK
            "cafe", "cafeteria" -> PlaceCategoryDto.CAFE
            "restaurant", "bar_restaurante" -> PlaceCategoryDto.RESTAURANT
            else -> PlaceCategoryDto.OTHER
        }
    }

    private fun String?.toFullMediaUrl(): String? {
        if (this.isNullOrBlank()) return null
        if (this.startsWith("http://") || this.startsWith("https://")) return this
        val rootHost = BuildConfig.BASE_URL.substringBefore("/api/v1")
        return if (this.startsWith("/")) {
            "$rootHost$this"
        } else {
            "$rootHost/$this"
        }
    }
}
