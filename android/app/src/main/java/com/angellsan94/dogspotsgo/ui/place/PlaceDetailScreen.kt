package com.angellsan94.dogspotsgo.ui.place

import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import coil.compose.AsyncImage
import com.angellsan94.dogspotsgo.domain.model.Comment
import com.angellsan94.dogspotsgo.domain.model.Photo

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PlaceDetailScreen(
    viewModel: PlaceDetailViewModel,
    onBack: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsState()
    var commentText by remember { mutableStateOf("") }
    var photoViewerIndex by remember { mutableIntStateOf(0) }
    var showPhotoViewer by remember { mutableStateOf(false) }
    val context = LocalContext.current
    val photoPicker = rememberLauncherForActivityResult(ActivityResultContracts.GetContent()) { uri ->
        if (uri != null) {
            try {
                val mimeType = context.contentResolver.getType(uri) ?: "image/jpeg"
                val image = context.contentResolver.openInputStream(uri)?.use { it.readBytes() }
                if (image == null) {
                    viewModel.showPhotoUploadError("No se pudo leer la imagen seleccionada.")
                } else {
                    viewModel.uploadPhoto(image, "photo", mimeType)
                }
            } catch (_: Exception) {
                viewModel.showPhotoUploadError("No se pudo leer la imagen seleccionada.")
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        text = (uiState as? PlaceDetailUiState.Success)?.place?.title ?: "Lugar"
                    )
                },
                navigationIcon = {
                    TextButton(onClick = onBack) {
                        Text("< Volver")
                    }
                }
            )
        }
    ) { paddingValues ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues),
            contentAlignment = Alignment.Center
        ) {
            when (val state = uiState) {
                is PlaceDetailUiState.Loading -> {
                    CircularProgressIndicator()
                }
                is PlaceDetailUiState.Error -> {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(text = state.message, color = MaterialTheme.colorScheme.error)
                        Spacer(modifier = Modifier.height(16.dp))
                        Button(onClick = { viewModel.loadPlaceDetailAndComments() }) {
                            Text("Reintentar")
                        }
                    }
                }
                is PlaceDetailUiState.Success -> {
                    val place = state.place
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .verticalScroll(rememberScrollState())
                    ) {
                        if (place.photos.isNotEmpty()) {
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(260.dp)
                            ) {
                                LazyRow(
                                    horizontalArrangement = Arrangement.spacedBy(4.dp),
                                    modifier = Modifier.fillMaxSize()
                                ) {
                                    itemsIndexed(place.photos, key = { _, photo -> photo.id }) { index, photo ->
                                        AsyncImage(
                                            model = photo.url,
                                            contentDescription = photo.caption ?: place.title,
                                            modifier = Modifier
                                                .fillMaxHeight()
                                                .width(320.dp)
                                                .clickable {
                                                    photoViewerIndex = index
                                                    showPhotoViewer = true
                                                },
                                            contentScale = ContentScale.Crop
                                        )
                                    }
                                }
                                Surface(
                                    modifier = Modifier
                                        .align(Alignment.BottomEnd)
                                        .padding(12.dp),
                                    color = Color.Black.copy(alpha = 0.7f),
                                    shape = MaterialTheme.shapes.small
                                ) {
                                    Text(
                                        text = "${place.photos.size} fotos",
                                        color = Color.White,
                                        style = MaterialTheme.typography.labelMedium,
                                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 7.dp)
                                    )
                                }
                            }
                        } else {
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(220.dp)
                                    .background(MaterialTheme.colorScheme.surfaceVariant),
                                contentAlignment = Alignment.Center
                            ) {
                                Text("Sin fotos")
                            }
                        }

                        Column(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 20.dp, vertical = 20.dp)
                        ) {
                            Text(
                                text = place.title,
                                style = MaterialTheme.typography.headlineLarge,
                                fontWeight = FontWeight.SemiBold
                            )
                            Spacer(modifier = Modifier.height(6.dp))
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(8.dp)
                            ) {
                                Text(place.category.name.replace('_', ' ').lowercase())
                                Text("·", color = MaterialTheme.colorScheme.outline)
                                Text("Por ${place.authorName}")
                            }
                            Spacer(modifier = Modifier.height(10.dp))
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(6.dp)
                            ) {
                                if (place.ratingAvg != null) {
                                    Text(
                                        text = "${String.format(java.util.Locale.getDefault(), "%.1f", place.ratingAvg)} ★",
                                        fontWeight = FontWeight.SemiBold
                                    )
                                    Text("(${place.ratingCount})", color = MaterialTheme.colorScheme.outline)
                                } else {
                                    Text("Aún sin valoraciones", color = MaterialTheme.colorScheme.outline)
                                }
                                Text("·", color = MaterialTheme.colorScheme.outline)
                                Text("${state.comments.size} reseñas")
                            }
                            Spacer(modifier = Modifier.height(18.dp))
                            if (state.isOwner) {
                                Button(
                                    onClick = { photoPicker.launch("image/*") },
                                    enabled = !state.isUploadingPhoto && place.photos.size < 8,
                                    modifier = Modifier.fillMaxWidth()
                                ) {
                                    Text(
                                        if (place.photos.size >= 8) "Límite de 8 fotos alcanzado"
                                        else "Añadir fotos (${place.photos.size}/8)"
                                    )
                                }
                            }
                            if (state.isUploadingPhoto) {
                                Row(
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    CircularProgressIndicator(modifier = Modifier.size(20.dp))
                                    Text("Subiendo foto...")
                                }
                            }
                            state.photoUploadError?.let { error ->
                                Text(error, color = MaterialTheme.colorScheme.error)
                            }
                            HorizontalDivider(modifier = Modifier.padding(vertical = 18.dp))

                            Text(text = "Valorar este lugar", style = MaterialTheme.typography.titleLarge)
                            Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                                for (i in 1..5) {
                                    TextButton(onClick = { viewModel.ratePlace(i) }) {
                                        Text(
                                            text = if (place.myRating != null && place.myRating >= i) "★" else "☆",
                                            style = MaterialTheme.typography.headlineSmall
                                        )
                                    }
                                }
                                if (place.myRating != null) {
                                    TextButton(onClick = { viewModel.deleteRating() }) {
                                        Text("Retirar", color = MaterialTheme.colorScheme.error)
                                    }
                                }
                            }

                        if (!place.description.isNullOrBlank()) {
                            Spacer(modifier = Modifier.height(16.dp))
                            Text(text = "Descripción", style = MaterialTheme.typography.titleMedium)
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(text = place.description, style = MaterialTheme.typography.bodyLarge)
                        }

                        if (place.tags.isNotEmpty()) {
                            Spacer(modifier = Modifier.height(16.dp))
                            Text(text = "Etiquetas", style = MaterialTheme.typography.titleMedium)
                            Spacer(modifier = Modifier.height(8.dp))
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.spacedBy(8.dp)
                            ) {
                                place.tags.forEach { tag ->
                                    AssistChip(
                                        onClick = {},
                                        label = { Text(tag.label) }
                                    )
                                }
                            }
                        }

                        // Comments Section
                        Spacer(modifier = Modifier.height(24.dp))
                        Text(text = "Comentarios", style = MaterialTheme.typography.titleLarge)
                        Spacer(modifier = Modifier.height(8.dp))

                        if (state.comments.isEmpty()) {
                            Text("No hay comentarios todavía. ¡Sé el primero en comentar!")
                        } else {
                            state.comments.forEach { comment ->
                                CommentItem(
                                    comment = comment,
                                    replies = state.repliesMap[comment.id] ?: emptyList(),
                                    onReplyClick = { viewModel.setReplyingTo(comment) },
                                    onLoadReplies = { viewModel.loadReplies(comment.id) },
                                    onDeleteClick = { viewModel.deleteComment(comment.id) },
                                    onReportClick = { viewModel.reportComment(comment.id, "inappropriate", null) }
                                )
                                Spacer(modifier = Modifier.height(8.dp))
                            }
                        }

                        // Comment input
                        Spacer(modifier = Modifier.height(16.dp))
                        if (state.replyingTo != null) {
                            Row(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalArrangement = Arrangement.SpaceBetween,
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Text(
                                    text = "Respondiendo a @${state.replyingTo.authorName}",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.primary
                                )
                                TextButton(onClick = { viewModel.setReplyingTo(null) }) {
                                    Text("Cancelar")
                                }
                            }
                        }

                        OutlinedTextField(
                            value = commentText,
                            onValueChange = { commentText = it },
                            label = { Text(if (state.replyingTo != null) "Escribe una respuesta..." else "Escribe un comentario...") },
                            modifier = Modifier.fillMaxWidth()
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Button(
                            onClick = {
                                viewModel.createComment(commentText, state.replyingTo?.id)
                                commentText = ""
                            },
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text("Enviar Comentario")
                        }
                    }
                    }
                }
            }

            val photos = (uiState as? PlaceDetailUiState.Success)?.place?.photos.orEmpty()
            if (showPhotoViewer && photos.isNotEmpty()) {
                PlacePhotoViewer(
                    photos = photos,
                    initialIndex = photoViewerIndex,
                    onDismiss = { showPhotoViewer = false }
                )
            }
        }
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun PlacePhotoViewer(
    photos: List<Photo>,
    initialIndex: Int,
    onDismiss: () -> Unit
) {
    val pagerState = rememberPagerState(
        initialPage = initialIndex.coerceIn(0, photos.lastIndex),
        pageCount = { photos.size }
    )

    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(usePlatformDefaultWidth = false)
    ) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black)
        ) {
            HorizontalPager(
                state = pagerState,
                modifier = Modifier.fillMaxSize()
            ) { page ->
                AsyncImage(
                    model = photos[page].url,
                    contentDescription = photos[page].caption,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Fit
                )
            }

            TextButton(
                onClick = onDismiss,
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .padding(16.dp)
            ) {
                Text("Cerrar", color = Color.White)
            }

            Surface(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(bottom = 24.dp),
                color = Color.Black.copy(alpha = 0.65f),
                shape = MaterialTheme.shapes.small
            ) {
                Text(
                    text = "${pagerState.currentPage + 1} / ${photos.size}",
                    color = Color.White,
                    style = MaterialTheme.typography.labelLarge,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
                )
            }
        }
    }
}

@Composable
fun CommentItem(
    comment: Comment,
    replies: List<Comment>,
    onReplyClick: () -> Unit,
    onLoadReplies: () -> Unit,
    onDeleteClick: () -> Unit,
    onReportClick: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(text = comment.authorName, style = MaterialTheme.typography.labelLarge)
                Text(text = comment.createdAt.take(10), style = MaterialTheme.typography.labelSmall)
            }
            Spacer(modifier = Modifier.height(4.dp))
            if (comment.deleted) {
                Text(text = "Comentario eliminado", style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.outline)
            } else {
                if (comment.replyToAuthorName != null) {
                    Text(
                        text = "@${comment.replyToAuthorName} ",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
                Text(text = comment.body ?: "", style = MaterialTheme.typography.bodyMedium)
            }

            Spacer(modifier = Modifier.height(8.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(16.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                if (!comment.deleted) {
                    TextButton(onClick = onReplyClick) {
                        Text("Responder")
                    }
                }
                if (comment.replyCount > 0 && replies.isEmpty()) {
                    TextButton(onClick = onLoadReplies) {
                        Text("Ver ${comment.replyCount} respuestas")
                    }
                }
                TextButton(onClick = onReportClick) {
                    Text("Denunciar", color = MaterialTheme.colorScheme.error)
                }
                TextButton(onClick = onDeleteClick) {
                    Text("Borrar", color = MaterialTheme.colorScheme.error)
                }
            }

            if (replies.isNotEmpty()) {
                Spacer(modifier = Modifier.height(8.dp))
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(start = 16.dp),
                    verticalArrangement = Arrangement.spacedBy(4.dp)
                ) {
                    replies.forEach { reply ->
                        Card(
                            modifier = Modifier.fillMaxWidth(),
                            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)
                        ) {
                            Column(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(8.dp)
                            ) {
                                Text(text = reply.authorName, style = MaterialTheme.typography.labelMedium)
                                if (reply.replyToAuthorName != null) {
                                    Text(
                                        text = "@${reply.replyToAuthorName} ",
                                        style = MaterialTheme.typography.bodySmall,
                                        color = MaterialTheme.colorScheme.primary
                                    )
                                }
                                Text(text = reply.body ?: "", style = MaterialTheme.typography.bodySmall)
                            }
                        }
                    }
                }
            }
        }
    }
}
