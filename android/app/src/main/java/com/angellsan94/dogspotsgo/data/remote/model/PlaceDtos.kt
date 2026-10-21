package com.angellsan94.dogspotsgo.data.remote.model

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
enum class PlaceCategoryDto {
    @SerialName("parque") PARQUE,
    @SerialName("bar_restaurante") BAR_RESTAURANTE,
    @SerialName("playa") PLAYA,
    @SerialName("ruta_natural") RUTA_NATURAL,
    @SerialName("pipican_area_canina") PIPICAN_AREA_CANINA,
    @SerialName("veterinario_tienda") VETERINARIO_TIENDA,
    @SerialName("otro") OTRO
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
