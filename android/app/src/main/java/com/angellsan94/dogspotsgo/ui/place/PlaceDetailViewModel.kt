package com.angellsan94.dogspotsgo.ui.place

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.angellsan94.dogspotsgo.domain.model.Comment
import com.angellsan94.dogspotsgo.domain.model.PlaceDetail
import com.angellsan94.dogspotsgo.domain.repository.CommentRepository
import com.angellsan94.dogspotsgo.domain.repository.PlaceRepository
import com.angellsan94.dogspotsgo.domain.repository.UserRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

sealed interface PlaceDetailUiState {
    data object Loading : PlaceDetailUiState
    data class Success(
        val place: PlaceDetail,
        val comments: List<Comment> = emptyList(),
        val repliesMap: Map<String, List<Comment>> = emptyMap(),
        val replyingTo: Comment? = null,
        val isOwner: Boolean = false,
        val isUploadingPhoto: Boolean = false,
        val photoUploadError: String? = null
    ) : PlaceDetailUiState
    data class Error(val message: String) : PlaceDetailUiState
}

@HiltViewModel
class PlaceDetailViewModel @Inject constructor(
    private val placeRepository: PlaceRepository,
    private val commentRepository: CommentRepository,
    private val userRepository: UserRepository,
    savedStateHandle: SavedStateHandle
) : ViewModel() {

    private val placeId: String = checkNotNull(savedStateHandle["placeId"])

    private val _uiState = MutableStateFlow<PlaceDetailUiState>(PlaceDetailUiState.Loading)
    val uiState: StateFlow<PlaceDetailUiState> = _uiState.asStateFlow()

    init {
        loadPlaceDetailAndComments()
    }

    fun loadPlaceDetailAndComments() {
        viewModelScope.launch {
            _uiState.update { PlaceDetailUiState.Loading }
            try {
                val place = placeRepository.getPlaceDetail(placeId)
                val comments = commentRepository.getComments(placeId)
                val isOwner = try {
                    userRepository.getMyProfile().id == place.authorId
                } catch (_: Exception) {
                    false
                }
                _uiState.update {
                    PlaceDetailUiState.Success(place = place, comments = comments, isOwner = isOwner)
                }
            } catch (e: Exception) {
                _uiState.update { PlaceDetailUiState.Error(e.localizedMessage ?: "Error al cargar el detalle del lugar.") }
            }
        }
    }

    fun uploadPhoto(image: ByteArray, fileName: String, mimeType: String) {
        val currentState = _uiState.value as? PlaceDetailUiState.Success ?: return
        if (!currentState.isOwner || currentState.place.photos.size >= 8 || currentState.isUploadingPhoto) return

        viewModelScope.launch {
            _uiState.update { state ->
                (state as? PlaceDetailUiState.Success)?.copy(
                    isUploadingPhoto = true,
                    photoUploadError = null
                ) ?: state
            }
            try {
                val photo = placeRepository.uploadPhoto(placeId, image, fileName, mimeType)
                _uiState.update { state ->
                    (state as? PlaceDetailUiState.Success)?.copy(
                        place = state.place.copy(photos = state.place.photos + photo),
                        isUploadingPhoto = false
                    ) ?: state
                }
            } catch (e: Exception) {
                _uiState.update { state ->
                    (state as? PlaceDetailUiState.Success)?.copy(
                        isUploadingPhoto = false,
                        photoUploadError = e.localizedMessage ?: "No se pudo subir la foto."
                    ) ?: state
                }
            }
        }
    }

    fun showPhotoUploadError(message: String) {
        _uiState.update { state ->
            (state as? PlaceDetailUiState.Success)?.copy(photoUploadError = message) ?: state
        }
    }

    fun ratePlace(rating: Int) {
        viewModelScope.launch {
            try {
                val summary = commentRepository.ratePlace(placeId, rating)
                val currentState = _uiState.value
                if (currentState is PlaceDetailUiState.Success) {
                    val updatedPlace = currentState.place.copy(
                        ratingAvg = summary.ratingAvg,
                        ratingCount = summary.ratingCount,
                        myRating = summary.myRating
                    )
                    _uiState.update { currentState.copy(place = updatedPlace) }
                }
            } catch (_: Exception) {}
        }
    }

    fun deleteRating() {
        viewModelScope.launch {
            try {
                val summary = commentRepository.deleteRating(placeId)
                val currentState = _uiState.value
                if (currentState is PlaceDetailUiState.Success) {
                    val updatedPlace = currentState.place.copy(
                        ratingAvg = summary.ratingAvg,
                        ratingCount = summary.ratingCount,
                        myRating = summary.myRating
                    )
                    _uiState.update { currentState.copy(place = updatedPlace) }
                }
            } catch (_: Exception) {}
        }
    }

    fun setReplyingTo(comment: Comment?) {
        val currentState = _uiState.value
        if (currentState is PlaceDetailUiState.Success) {
            _uiState.update { currentState.copy(replyingTo = comment) }
        }
    }

    fun createComment(body: String, replyToCommentId: String?) {
        if (body.isBlank()) return
        viewModelScope.launch {
            try {
                commentRepository.createComment(placeId, body.trim(), replyToCommentId)
                val comments = commentRepository.getComments(placeId)
                val currentState = _uiState.value
                if (currentState is PlaceDetailUiState.Success) {
                    _uiState.update { currentState.copy(comments = comments, replyingTo = null) }
                }
            } catch (_: Exception) {}
        }
    }

    fun loadReplies(commentId: String) {
        viewModelScope.launch {
            try {
                val replies = commentRepository.getReplies(commentId)
                val currentState = _uiState.value
                if (currentState is PlaceDetailUiState.Success) {
                    val updatedMap = currentState.repliesMap.toMutableMap()
                    updatedMap[commentId] = replies
                    _uiState.update { currentState.copy(repliesMap = updatedMap) }
                }
            } catch (_: Exception) {}
        }
    }

    fun deleteComment(commentId: String) {
        viewModelScope.launch {
            try {
                commentRepository.deleteComment(commentId)
                val comments = commentRepository.getComments(placeId)
                val currentState = _uiState.value
                if (currentState is PlaceDetailUiState.Success) {
                    _uiState.update { currentState.copy(comments = comments) }
                }
            } catch (_: Exception) {}
        }
    }

    fun reportPlace(reason: String, details: String?) {
        viewModelScope.launch {
            try {
                commentRepository.reportPlace(placeId, reason, details)
            } catch (_: Exception) {}
        }
    }

    fun reportComment(commentId: String, reason: String, details: String?) {
        viewModelScope.launch {
            try {
                commentRepository.reportComment(commentId, reason, details)
            } catch (_: Exception) {}
        }
    }
}
