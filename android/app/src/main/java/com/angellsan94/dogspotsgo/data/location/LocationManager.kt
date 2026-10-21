package com.angellsan94.dogspotsgo.data.location

import android.annotation.SuppressLint
import android.content.Context
import android.location.Location
import com.google.android.gms.location.Priority
import com.google.android.gms.location.FusedLocationProviderClient
import com.google.android.gms.location.LocationServices
import com.google.android.gms.tasks.CancellationTokenSource
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.suspendCancellableCoroutine
import javax.inject.Inject
import javax.inject.Singleton
import kotlin.coroutines.resume

@Singleton
class LocationManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val fusedLocationClient: FusedLocationProviderClient =
        LocationServices.getFusedLocationProviderClient(context)

    @SuppressLint("MissingPermission")
    suspend fun getCurrentLocation(): Location? = suspendCancellableCoroutine { continuation ->
        val cancellationTokenSource = CancellationTokenSource()
        continuation.invokeOnCancellation { cancellationTokenSource.cancel() }

        fun complete(location: Location?) {
            if (continuation.isActive) continuation.resume(location)
        }

        fun requestLastKnownLocation() {
            try {
                fusedLocationClient.lastLocation
                    .addOnSuccessListener(::complete)
                    .addOnFailureListener { complete(null) }
            } catch (_: Exception) {
                complete(null)
            }
        }

        try {
            fusedLocationClient.getCurrentLocation(
                Priority.PRIORITY_BALANCED_POWER_ACCURACY,
                cancellationTokenSource.token
            ).addOnSuccessListener { location ->
                if (location != null) complete(location) else requestLastKnownLocation()
            }.addOnFailureListener { requestLastKnownLocation() }
        } catch (e: Exception) {
            requestLastKnownLocation()
        }
    }
}
