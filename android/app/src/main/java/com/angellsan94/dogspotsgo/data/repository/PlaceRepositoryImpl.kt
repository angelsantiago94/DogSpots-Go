package com.angellsan94.dogspotsgo.data.repository

import com.angellsan94.dogspotsgo.data.remote.PlacesApiService
import com.angellsan94.dogspotsgo.data.remote.model.PlaceCategoryDto
import com.angellsan94.dogspotsgo.domain.model.Place
import com.angellsan94.dogspotsgo.domain.model.PlaceCategory
import com.angellsan94.dogspotsgo.domain.repository.PlaceRepository
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
    ): List<Place> {
        val response = placesApiService.getPlaces(
            minLat = minLat,
            minLng = minLng,
            maxLat = maxLat,
            maxLng = maxLng,
            category = category,
            minRating = minRating
        )
        return response.items.map { dto ->
            Place(
                id = dto.id,
                title = dto.title,
                lat = dto.lat,
                lng = dto.lng,
                category = when (dto.category) {
                    PlaceCategoryDto.PARQUE -> PlaceCategory.PARQUE
                    PlaceCategoryDto.BAR_RESTAURANTE -> PlaceCategory.BAR_RESTAURANTE
                    PlaceCategoryDto.PLAYA -> PlaceCategory.PLAYA
                    PlaceCategoryDto.RUTA_NATURAL -> PlaceCategory.RUTA_NATURAL
                    PlaceCategoryDto.PIPICAN_AREA_CANINA -> PlaceCategory.PIPICAN_AREA_CANINA
                    PlaceCategoryDto.VETERINARIO_TIENDA -> PlaceCategory.VETERINARIO_TIENDA
                    PlaceCategoryDto.OTRO -> PlaceCategory.OTRO
                },
                coverThumbUrl = dto.cover_thumb_url,
                ratingAvg = dto.rating_avg,
                ratingCount = dto.rating_count,
                distanceM = dto.distance_m
            )
        }
    }
}
