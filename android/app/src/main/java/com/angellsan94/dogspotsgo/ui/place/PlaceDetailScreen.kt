package com.angellsan94.dogspotsgo.ui.place

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import com.angellsan94.dogspotsgo.domain.model.Comment

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PlaceDetailScreen(
    viewModel: PlaceDetailViewModel,
    onBack: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsState()
    var commentText by remember { mutableStateOf("") }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Detalle del Lugar") },
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
                            .padding(16.dp)
                    ) {
                        if (place.photos.isNotEmpty()) {
                            LazyRow(
                                horizontalArrangement = Arrangement.spacedBy(8.dp),
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(200.dp)
                            ) {
                                items(place.photos) { photo ->
                                    AsyncImage(
                                        model = photo.url,
                                        contentDescription = photo.caption,
                                        modifier = Modifier
                                            .fillMaxHeight()
                                            .aspectRatio(1f),
                                        contentScale = ContentScale.Crop
                                    )
                                }
                            }
                            Spacer(modifier = Modifier.height(16.dp))
                        }

                        Text(text = place.title, style = MaterialTheme.typography.headlineMedium)
                        Spacer(modifier = Modifier.height(4.dp))
                        Text(text = "Creado por: ${place.authorName}", style = MaterialTheme.typography.bodySmall)
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(text = "Categoría: ${place.category.name}", style = MaterialTheme.typography.bodyMedium)

                        if (place.ratingAvg != null) {
                            Spacer(modifier = Modifier.height(4.dp))
                            Text(text = "Valoración media: ${place.ratingAvg} / 5 (${place.ratingCount} opiniones)", style = MaterialTheme.typography.bodyMedium)
                        }

                        // Rating selector (1-5 stars)
                        Spacer(modifier = Modifier.height(16.dp))
                        Text(text = "Tu valoración:", style = MaterialTheme.typography.titleMedium)
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
