package com.recipes.mobile.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

private val LightColors = lightColorScheme(
    primary = Color(0xFF1B64B0),
    onPrimary = Color(0xFFF8FBFF),
    primaryContainer = Color(0xFFDCEBFF),
    onPrimaryContainer = Color(0xFF09284A),
    secondary = Color(0xFF3A5F86),
    background = Color(0xFFF5F8FD),
    surface = Color(0xFFFCFDFF),
)

private val DarkColors = darkColorScheme(
    primary = Color(0xFFA8CAFF),
    onPrimary = Color(0xFF002A52),
)

@Composable
fun RecipesTheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colorScheme = LightColors,
        typography = Typography,
        content = content
    )
}
