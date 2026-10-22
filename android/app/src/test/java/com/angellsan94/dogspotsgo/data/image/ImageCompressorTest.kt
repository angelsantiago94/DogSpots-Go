package com.angellsan94.dogspotsgo.data.image

import org.junit.Assert.assertEquals
import org.junit.Test

class ImageCompressorTest {

    @Test
    fun sampleSizeDoesNotDownscaleSmallImages() {
        assertEquals(1, sampleSize(1920, 1080, 2048))
    }

    @Test
    fun sampleSizeHalvesLargeImages() {
        assertEquals(2, sampleSize(8000, 6000, 2048))
    }

    @Test
    fun sampleSizeRepeatsForVeryLargeImages() {
        assertEquals(4, sampleSize(12000, 9000, 2048))
    }
}
