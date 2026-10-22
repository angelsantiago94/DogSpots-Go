package com.angellsan94.dogspotsgo.data.image

import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.graphics.Canvas
import android.graphics.Color
import android.graphics.Matrix
import android.media.ExifInterface
import java.io.ByteArrayInputStream
import java.io.ByteArrayOutputStream
import kotlin.math.max

private const val MAX_DIMENSION = 2048
private const val JPEG_QUALITY = 85

// Reduce la imagen a 2048px y JPEG de calidad 85 antes de subirla.
// Si no se puede decodificar, devuelve los bytes originales para que el servidor valide.
@Suppress("DEPRECATION")
fun compressToJpeg(bytes: ByteArray): ByteArray {
    val bounds = BitmapFactory.Options().apply { inJustDecodeBounds = true }
    BitmapFactory.decodeByteArray(bytes, 0, bytes.size, bounds)
    if (bounds.outWidth <= 0 || bounds.outHeight <= 0) return bytes
    val sourceHasAlpha = bounds.outMimeType == "image/png" ||
        bounds.outMimeType == "image/webp" ||
        bounds.outMimeType == "image/gif"

    val options = BitmapFactory.Options().apply {
        inSampleSize = sampleSize(bounds.outWidth, bounds.outHeight, MAX_DIMENSION)
    }
    val decoded = BitmapFactory.decodeByteArray(bytes, 0, bytes.size, options) ?: return bytes
    val orientation = runCatching {
        ExifInterface(ByteArrayInputStream(bytes))
            .getAttributeInt(ExifInterface.TAG_ORIENTATION, ExifInterface.ORIENTATION_NORMAL)
    }.getOrDefault(ExifInterface.ORIENTATION_NORMAL)

    val longestSide = max(decoded.width, decoded.height)
    val scaled = if (longestSide > MAX_DIMENSION) {
        val ratio = MAX_DIMENSION.toFloat() / longestSide
        Bitmap.createScaledBitmap(
            decoded,
            (decoded.width * ratio).toInt().coerceAtLeast(1),
            (decoded.height * ratio).toInt().coerceAtLeast(1),
            true
        )
    } else {
        decoded
    }

    val oriented = applyOrientation(scaled, orientation)
    // JPEG no soporta transparencia: PNG/WebP con alfa se aplanan sobre blanco.
    val flattened = if (sourceHasAlpha && oriented.hasAlpha) {
        Bitmap.createBitmap(oriented.width, oriented.height, Bitmap.Config.ARGB_8888).also { bitmap ->
            Canvas(bitmap).apply {
                drawColor(Color.WHITE)
                drawBitmap(oriented, 0f, 0f, null)
            }
        }
    } else {
        oriented
    }
    val output = ByteArrayOutputStream()
    try {
        flattened.compress(Bitmap.CompressFormat.JPEG, JPEG_QUALITY, output)
    } finally {
        if (flattened !== oriented) flattened.recycle()
        if (oriented !== scaled) oriented.recycle()
        if (scaled !== decoded) scaled.recycle()
        decoded.recycle()
    }
    return output.toByteArray()
}

internal fun sampleSize(width: Int, height: Int, maxDimension: Int): Int {
    var sample = 1
    while (max(width, height) / (sample * 2) >= maxDimension) {
        sample *= 2
    }
    return sample
}

@Suppress("DEPRECATION")
private fun applyOrientation(bitmap: Bitmap, orientation: Int): Bitmap {
    val matrix = Matrix()
    when (orientation) {
        ExifInterface.ORIENTATION_ROTATE_90 -> matrix.postRotate(90f)
        ExifInterface.ORIENTATION_ROTATE_180 -> matrix.postRotate(180f)
        ExifInterface.ORIENTATION_ROTATE_270 -> matrix.postRotate(270f)
        ExifInterface.ORIENTATION_FLIP_HORIZONTAL -> matrix.postScale(-1f, 1f)
        ExifInterface.ORIENTATION_FLIP_VERTICAL -> matrix.postScale(1f, -1f)
        ExifInterface.ORIENTATION_TRANSPOSE -> {
            matrix.postRotate(90f)
            matrix.postScale(-1f, 1f)
        }
        ExifInterface.ORIENTATION_TRANSVERSE -> {
            matrix.postRotate(270f)
            matrix.postScale(-1f, 1f)
        }
        else -> return bitmap
    }
    return Bitmap.createBitmap(bitmap, 0, 0, bitmap.width, bitmap.height, matrix, true)
}
