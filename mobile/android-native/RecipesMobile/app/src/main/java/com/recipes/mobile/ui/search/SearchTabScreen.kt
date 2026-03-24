package com.recipes.mobile.ui.search

import android.Manifest
import android.app.Activity
import android.content.Intent
import android.content.pm.PackageManager
import android.speech.RecognizerIntent
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.FilterChip
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Switch
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalView
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import com.recipes.mobile.BuildConfig
import com.recipes.mobile.core.network.RecipesApiClient
import com.recipes.mobile.core.network.SearchPayload
import com.recipes.mobile.core.voice.VoiceTelemetryStore
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.util.Locale

@Composable
fun SearchTabScreen() {
    val context = LocalContext.current
    val view = LocalView.current
    var ingredientsRaw by remember { mutableStateOf("chicken, rice, garlic") }
    var mode by remember { mutableStateOf("strict") }
    var complex by remember { mutableStateOf(false) }
    var loading by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }
    var payload by remember { mutableStateOf<SearchPayload?>(null) }
    var voiceError by remember { mutableStateOf<String?>(null) }
    val scope = rememberCoroutineScope()

    val client = remember { RecipesApiClient(BuildConfig.API_BASE_URL) }
    val voiceTelemetry = remember { VoiceTelemetryStore(context) }

    val speechLauncher = rememberLauncherForActivityResult(ActivityResultContracts.StartActivityForResult()) { result ->
        if (result.resultCode != Activity.RESULT_OK) {
            voiceError = "Voice input canceled. Tap Speak ingredients to try again."
            voiceTelemetry.incrementFailure()
            return@rememberLauncherForActivityResult
        }

        val transcript = result.data
            ?.getStringArrayListExtra(RecognizerIntent.EXTRA_RESULTS)
            ?.firstOrNull()
            ?.trim()

        if (transcript.isNullOrBlank()) {
            voiceError = "No speech recognized. Try speaking clearly and closer to the microphone."
            voiceTelemetry.incrementFailure()
            return@rememberLauncherForActivityResult
        }

        ingredientsRaw = mergeIngredients(ingredientsRaw, transcript)
        voiceError = null
        voiceTelemetry.incrementSuccess()
        view.announceForAccessibility("Voice input added")
    }

    fun launchSpeechRecognition() {
        val intent = Intent(RecognizerIntent.ACTION_RECOGNIZE_SPEECH).apply {
            putExtra(RecognizerIntent.EXTRA_LANGUAGE_MODEL, RecognizerIntent.LANGUAGE_MODEL_FREE_FORM)
            putExtra(RecognizerIntent.EXTRA_LANGUAGE, Locale.getDefault())
            putExtra(RecognizerIntent.EXTRA_PROMPT, "Speak ingredients")
            putExtra(RecognizerIntent.EXTRA_PARTIAL_RESULTS, false)
        }

        if (intent.resolveActivity(context.packageManager) == null) {
            voiceError = "Speech recognition unavailable on this device. Use manual typing for ingredients."
            voiceTelemetry.incrementFailure()
            return
        }

        voiceTelemetry.incrementStart()
        speechLauncher.launch(intent)
    }

    val micPermissionLauncher = rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) { granted ->
        if (granted) {
            launchSpeechRecognition()
            return@rememberLauncherForActivityResult
        }
        voiceError = "Microphone permission denied. You can continue with typed ingredients."
        voiceTelemetry.incrementPermissionDenied()
    }

    LaunchedEffect(error) {
        if (!error.isNullOrBlank()) {
            view.announceForAccessibility("Search error. $error")
        }
    }

    LaunchedEffect(voiceError) {
        if (!voiceError.isNullOrBlank()) {
            view.announceForAccessibility(voiceError)
        }
    }

    LaunchedEffect(payload?.total) {
        val total = payload?.total ?: return@LaunchedEffect
        view.announceForAccessibility("Search completed. $total results.")
    }

    suspend fun runSearch() {
        loading = true
        error = null
        try {
            val ingredients = ingredientsRaw
                .split(',')
                .map { it.trim().lowercase() }
                .filter { it.isNotBlank() }
                .distinct()

            payload = withContext(Dispatchers.IO) {
                client.search(ingredients = ingredients, mode = mode, complex = complex)
            }
        } catch (e: Exception) {
            error = e.message ?: "Search failed"
        } finally {
            loading = false
        }
    }

    LazyColumn(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        item {
            Text("Discover Recipes", modifier = Modifier.semantics { heading() })
        }
        item {
            Text("Use pantry ingredients to get ranked recipe matches.")
        }
        item {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(
                    modifier = Modifier
                        .fillMaxWidth()
                        .semantics { contentDescription = "Ingredients input" },
                    value = ingredientsRaw,
                    onValueChange = { ingredientsRaw = it },
                    label = { Text("Ingredients (comma-separated)") }
                )

                Button(
                    onClick = {
                        val hasMicPermission = ContextCompat.checkSelfPermission(
                            context,
                            Manifest.permission.RECORD_AUDIO
                        ) == PackageManager.PERMISSION_GRANTED

                        if (hasMicPermission) {
                            launchSpeechRecognition()
                        } else {
                            micPermissionLauncher.launch(Manifest.permission.RECORD_AUDIO)
                        }
                    },
                    modifier = Modifier.semantics { contentDescription = "Dictate ingredients" }
                ) {
                    Text("Dictate ingredients")
                }
            }
        }
        item {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                FilterChip(
                    selected = mode == "strict",
                    onClick = { mode = "strict" },
                    label = { Text("Strict") },
                    modifier = Modifier.semantics { contentDescription = "Strict mode" }
                )
                FilterChip(
                    selected = mode == "inclusive",
                    onClick = { mode = "inclusive" },
                    label = { Text("Inclusive") },
                    modifier = Modifier.semantics { contentDescription = "Inclusive mode" }
                )
            }
        }
        item {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Text("Complex")
                Switch(
                    checked = complex,
                    onCheckedChange = { complex = it },
                    modifier = Modifier.semantics { contentDescription = "Complex mode toggle" }
                )
            }
        }
        item {
            Button(
                onClick = { scope.launch { runSearch() } },
                enabled = !loading,
                modifier = Modifier.semantics { contentDescription = "Run recipe search" }
            ) {
                Text(if (loading) "Finding matches..." else "Find matches")
            }
        }

        if (error != null) {
            item {
                Card { Text("Error: $error", modifier = Modifier.padding(12.dp)) }
            }
        }

        if (voiceError != null) {
            item {
                Card { Text("Voice: $voiceError", modifier = Modifier.padding(12.dp)) }
            }
        }

        payload?.let { data ->
            item {
                Text("Mode: ${data.mode} | Total: ${data.total}")
            }
            items(data.results.take(20)) { item ->
                Card(modifier = Modifier.fillMaxWidth()) {
                    Column(modifier = Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                        Text(item.name)
                        Text("${item.source} • ${(item.matchPercent * 100).toInt()}% match")
                    }
                }
            }
        }
    }
}

private fun mergeIngredients(existing: String, transcript: String): String {
    val normalized = transcript
        .replace(" and ", ", ")
        .replace(';', ',')
        .trim()

    if (existing.isBlank()) {
        return normalized
    }

    val suffix = if (existing.trimEnd().endsWith(',')) " " else ", "
    return existing.trim() + suffix + normalized
}
