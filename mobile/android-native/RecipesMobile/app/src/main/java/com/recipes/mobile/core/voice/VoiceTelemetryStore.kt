package com.recipes.mobile.core.voice

import android.content.Context

class VoiceTelemetryStore(context: Context) {
    private val prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    fun incrementStart() {
        increment(KEY_START)
    }

    fun incrementSuccess() {
        increment(KEY_SUCCESS)
    }

    fun incrementFailure() {
        increment(KEY_FAILURE)
    }

    fun incrementPermissionDenied() {
        increment(KEY_PERMISSION_DENIED)
    }

    private fun increment(key: String) {
        val next = prefs.getLong(key, 0L) + 1L
        prefs.edit().putLong(key, next).apply()
    }

    private companion object {
        const val PREFS_NAME = "recipes_voice_telemetry"
        const val KEY_START = "stt_start"
        const val KEY_SUCCESS = "stt_success"
        const val KEY_FAILURE = "stt_failure"
        const val KEY_PERMISSION_DENIED = "stt_permission_denied"
    }
}
