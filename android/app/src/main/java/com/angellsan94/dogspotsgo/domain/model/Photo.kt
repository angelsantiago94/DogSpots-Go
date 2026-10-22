package com.angellsan94.dogspotsgo.domain.model

data class Photo(
    val id: String,
    val url: String,
    val thumbUrl: String,
    val caption: String?,
    val width: Int,
    val height: Int,
    val position: Int,
    val uploadedBy: String,
    val createdAt: String
)
