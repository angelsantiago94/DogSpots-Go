package com.angellsan94.dogspotsgo.data.remote.model

import kotlinx.serialization.Serializable

@Serializable
data class UpdateUserRequestDto(
    val name: String? = null
)
