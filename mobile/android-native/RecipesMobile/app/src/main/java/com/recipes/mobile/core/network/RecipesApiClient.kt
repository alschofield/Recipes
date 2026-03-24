package com.recipes.mobile.core.network

import java.io.BufferedReader
import java.io.InputStreamReader
import java.net.HttpURLConnection
import java.net.URL
import org.json.JSONArray
import org.json.JSONObject

class RecipesApiClient(private val baseUrl: String) {
    fun search(ingredients: List<String>, mode: String, complex: Boolean): SearchPayload {
        val payload = JSONObject().apply {
            put("ingredients", JSONArray(ingredients))
            put("mode", mode)
            put("complex", complex)
            put("dbOnly", false)
        }

        val url = URL("${baseUrl.trimEnd('/')}/recipes/search")
        val connection = (url.openConnection() as HttpURLConnection).apply {
            requestMethod = "POST"
            setRequestProperty("Content-Type", "application/json")
            connectTimeout = 10_000
            readTimeout = 20_000
            doOutput = true
        }

        connection.outputStream.use { it.write(payload.toString().toByteArray()) }

        val code = connection.responseCode
        val input = if (code in 200..299) connection.inputStream else connection.errorStream
        val raw = BufferedReader(InputStreamReader(input)).readText()

        if (code !in 200..299) {
            throw IllegalStateException("Search failed ($code): $raw")
        }

        val response = JSONObject(raw)
        val pagination = response.optJSONObject("pagination")
        val total = pagination?.optInt("total", 0) ?: 0
        val modeOut = response.optString("mode", mode)

        val resultsArray = response.optJSONArray("results") ?: JSONArray()
        val results = buildList {
            for (i in 0 until resultsArray.length()) {
                val item = resultsArray.optJSONObject(i) ?: continue
                add(
                    SearchRecipeItem(
                        id = item.optString("id", ""),
                        name = item.optString("name", ""),
                        source = item.optString("source", "unknown"),
                        matchPercent = item.optDouble("matchPercent", 0.0),
                    )
                )
            }
        }

        return SearchPayload(mode = modeOut, total = total, results = results)
    }

    fun listFavorites(userId: String, token: String): List<FavoriteItem> {
        val url = URL("${baseUrl.trimEnd('/')}/favorites/$userId")
        val connection = (url.openConnection() as HttpURLConnection).apply {
            requestMethod = "GET"
            setRequestProperty("Authorization", "Bearer $token")
            connectTimeout = 10_000
            readTimeout = 20_000
        }

        val code = connection.responseCode
        val input = if (code in 200..299) connection.inputStream else connection.errorStream
        val raw = BufferedReader(InputStreamReader(input)).readText()

        if (code !in 200..299) {
            throw IllegalStateException("Favorites fetch failed ($code): $raw")
        }

        val array = JSONArray(raw)
        return buildList {
            for (i in 0 until array.length()) {
                val item = array.optJSONObject(i) ?: continue
                add(
                    FavoriteItem(
                        id = item.optString("id", ""),
                        userId = item.optString("userId", ""),
                        recipeId = item.optString("recipeId", ""),
                        recipeName = item.optString("recipeName", null),
                    )
                )
            }
        }
    }

    fun addFavorite(userId: String, recipeId: String, token: String) {
        val url = URL("${baseUrl.trimEnd('/')}/favorites/$userId/$recipeId")
        val connection = (url.openConnection() as HttpURLConnection).apply {
            requestMethod = "POST"
            setRequestProperty("Authorization", "Bearer $token")
            connectTimeout = 10_000
            readTimeout = 20_000
            doOutput = true
        }

        val code = connection.responseCode
        if (code !in 200..299) {
            val raw = BufferedReader(InputStreamReader(connection.errorStream)).readText()
            throw IllegalStateException("Add favorite failed ($code): $raw")
        }
    }

    fun removeFavorite(userId: String, recipeId: String, token: String) {
        val url = URL("${baseUrl.trimEnd('/')}/favorites/$userId/$recipeId")
        val connection = (url.openConnection() as HttpURLConnection).apply {
            requestMethod = "DELETE"
            setRequestProperty("Authorization", "Bearer $token")
            connectTimeout = 10_000
            readTimeout = 20_000
        }

        val code = connection.responseCode
        if (code !in 200..299) {
            val raw = BufferedReader(InputStreamReader(connection.errorStream)).readText()
            throw IllegalStateException("Remove favorite failed ($code): $raw")
        }
    }

    fun login(username: String, password: String, clientSessionId: String): LoginResponse {
        val payload = JSONObject().apply {
            put("username", username)
            put("password", password)
        }

        val response = requestJson(
            path = "/users/login",
            method = "POST",
            token = null,
            body = payload,
            extraHeaders = mapOf("X-Client-Session-ID" to clientSessionId)
        )

        return LoginResponse(
            id = response.optString("id", ""),
            username = response.optString("username", ""),
            email = response.optString("email", ""),
            role = response.optString("role", "user"),
            token = response.optString("token", ""),
            expiresAt = response.optString("expiresAt", ""),
            refreshToken = response.optString("refreshToken", ""),
            refreshExpiresAt = response.optString("refreshExpiresAt", ""),
            sessionId = response.optString("sessionId", clientSessionId),
        )
    }

    fun refresh(refreshToken: String): RefreshResponse {
        val payload = JSONObject().apply { put("refreshToken", refreshToken) }
        val response = requestJson(path = "/users/refresh", method = "POST", token = null, body = payload)

        return RefreshResponse(
            token = response.optString("token", ""),
            expiresAt = response.optString("expiresAt", ""),
            refreshToken = response.optString("refreshToken", ""),
            refreshExpiresAt = response.optString("refreshExpiresAt", ""),
            sessionId = response.optString("sessionId", ""),
        )
    }

    fun logout(refreshToken: String) {
        val payload = JSONObject().apply { put("refreshToken", refreshToken) }
        requestNoContent(path = "/users/logout", method = "POST", token = null, body = payload)
    }

    fun logoutSession(refreshToken: String) {
        val payload = JSONObject().apply { put("refreshToken", refreshToken) }
        requestNoContent(path = "/users/logout/session", method = "POST", token = null, body = payload)
    }

    fun listSessions(userId: String, token: String): List<SessionItem> {
        val response = requestArray(path = "/users/$userId/sessions", method = "GET", token = token)
        return buildList {
            for (i in 0 until response.length()) {
                val item = response.optJSONObject(i) ?: continue
                add(
                    SessionItem(
                        sessionId = item.optString("sessionId", ""),
                        createdAt = item.optString("createdAt", ""),
                        lastUsedAt = item.optString("lastUsedAt", null),
                        userAgent = item.optString("userAgent", null),
                        ipAddress = item.optString("ipAddress", null),
                    )
                )
            }
        }
    }

    private fun requestJson(
        path: String,
        method: String,
        token: String?,
        body: JSONObject? = null,
        extraHeaders: Map<String, String> = emptyMap(),
    ): JSONObject {
        val connection = open(path = path, method = method, token = token, body = body, extraHeaders = extraHeaders)
        val (code, raw) = readResponse(connection)
        if (code !in 200..299) {
            throw IllegalStateException("Request failed ($code): $raw")
        }
        return JSONObject(raw)
    }

    private fun requestArray(path: String, method: String, token: String?): JSONArray {
        val connection = open(path = path, method = method, token = token)
        val (code, raw) = readResponse(connection)
        if (code !in 200..299) {
            throw IllegalStateException("Request failed ($code): $raw")
        }
        return JSONArray(raw)
    }

    private fun requestNoContent(path: String, method: String, token: String?, body: JSONObject? = null) {
        val connection = open(path = path, method = method, token = token, body = body)
        val (code, raw) = readResponse(connection)
        if (code !in 200..299) {
            throw IllegalStateException("Request failed ($code): $raw")
        }
    }

    private fun open(
        path: String,
        method: String,
        token: String?,
        body: JSONObject? = null,
        extraHeaders: Map<String, String> = emptyMap(),
    ): HttpURLConnection {
        val normalizedPath = path.trimStart('/')
        val url = URL("${baseUrl.trimEnd('/')}/$normalizedPath")
        val connection = (url.openConnection() as HttpURLConnection).apply {
            requestMethod = method
            connectTimeout = 10_000
            readTimeout = 20_000
            setRequestProperty("Accept", "application/json")
            if (token != null) {
                setRequestProperty("Authorization", "Bearer $token")
            }
            extraHeaders.forEach { (name, value) -> setRequestProperty(name, value) }
            if (body != null) {
                setRequestProperty("Content-Type", "application/json")
                doOutput = true
                outputStream.use { it.write(body.toString().toByteArray()) }
            }
        }
        return connection
    }

    private fun readResponse(connection: HttpURLConnection): Pair<Int, String> {
        val code = connection.responseCode
        val stream = if (code in 200..299) connection.inputStream else connection.errorStream
        val text = stream?.let { BufferedReader(InputStreamReader(it)).readText() } ?: ""
        return code to text
    }
}
