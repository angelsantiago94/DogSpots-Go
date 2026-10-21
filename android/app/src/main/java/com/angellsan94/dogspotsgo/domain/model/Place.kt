package com.angellsan94.dogspotsgo.domain.model

enum class PlaceCategory {
    PARQUE, BAR_RESTAURANTE, PLAYA, RUTA_NATURAL, PIPICAN_AREA_CANINA, VETERINARIO_TIENDA, OTRO
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
