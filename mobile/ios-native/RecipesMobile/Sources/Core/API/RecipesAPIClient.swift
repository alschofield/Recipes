import Foundation

enum RecipesAPIError: Error {
    case invalidResponse
    case serverError(status: Int, message: String)
}

private struct RefreshRequestPayload: Encodable {
    let refreshToken: String
}

private struct LoginRequestPayload: Encodable {
    let username: String
    let password: String
}

struct RecipesAPIClient {
    private let session: URLSession
    private let baseURL: URL

    init(baseURL: URL = AppConfig.apiBaseURL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
    }

    func search(ingredients: [String], mode: String, complex: Bool) async throws -> SearchResponse {
        let payload = SearchRequestPayload(ingredients: ingredients, mode: mode, complex: complex, dbOnly: false)

        var request = URLRequest(url: baseURL.appendingPathComponent("recipes/search"))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(payload)

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else {
            throw RecipesAPIError.invalidResponse
        }

        if !(200...299).contains(http.statusCode) {
            let message = String(data: data, encoding: .utf8) ?? "unknown"
            throw RecipesAPIError.serverError(status: http.statusCode, message: message)
        }

        return try JSONDecoder().decode(SearchResponse.self, from: data)
    }

    func listFavorites(userID: String, token: String) async throws -> [FavoriteItem] {
        var request = URLRequest(url: baseURL.appendingPathComponent("favorites/\(userID)"))
        request.httpMethod = "GET"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else {
            throw RecipesAPIError.invalidResponse
        }

        if !(200...299).contains(http.statusCode) {
            let message = String(data: data, encoding: .utf8) ?? "unknown"
            throw RecipesAPIError.serverError(status: http.statusCode, message: message)
        }

        return try JSONDecoder().decode([FavoriteItem].self, from: data)
    }

    func addFavorite(userID: String, recipeID: String, token: String) async throws {
        var request = URLRequest(url: baseURL.appendingPathComponent("favorites/\(userID)/\(recipeID)"))
        request.httpMethod = "POST"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else {
            throw RecipesAPIError.invalidResponse
        }

        if !(200...299).contains(http.statusCode) {
            let message = String(data: data, encoding: .utf8) ?? "unknown"
            throw RecipesAPIError.serverError(status: http.statusCode, message: message)
        }
    }

    func removeFavorite(userID: String, recipeID: String, token: String) async throws {
        var request = URLRequest(url: baseURL.appendingPathComponent("favorites/\(userID)/\(recipeID)"))
        request.httpMethod = "DELETE"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else {
            throw RecipesAPIError.invalidResponse
        }

        if !(200...299).contains(http.statusCode) {
            let message = String(data: data, encoding: .utf8) ?? "unknown"
            throw RecipesAPIError.serverError(status: http.statusCode, message: message)
        }
    }

    func login(username: String, password: String, clientSessionID: String) async throws -> LoginResponse {
        var request = URLRequest(url: baseURL.appendingPathComponent("users/login"))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue(clientSessionID, forHTTPHeaderField: "X-Client-Session-ID")
        request.httpBody = try JSONEncoder().encode(LoginRequestPayload(username: username, password: password))

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else {
            throw RecipesAPIError.invalidResponse
        }

        if !(200...299).contains(http.statusCode) {
            let message = String(data: data, encoding: .utf8) ?? "unknown"
            throw RecipesAPIError.serverError(status: http.statusCode, message: message)
        }

        return try JSONDecoder().decode(LoginResponse.self, from: data)
    }

    func refresh(refreshToken: String) async throws -> RefreshResponse {
        var request = URLRequest(url: baseURL.appendingPathComponent("users/refresh"))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(RefreshRequestPayload(refreshToken: refreshToken))

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else {
            throw RecipesAPIError.invalidResponse
        }

        if !(200...299).contains(http.statusCode) {
            let message = String(data: data, encoding: .utf8) ?? "unknown"
            throw RecipesAPIError.serverError(status: http.statusCode, message: message)
        }

        return try JSONDecoder().decode(RefreshResponse.self, from: data)
    }

    func logout(refreshToken: String) async throws {
        var request = URLRequest(url: baseURL.appendingPathComponent("users/logout"))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(RefreshRequestPayload(refreshToken: refreshToken))

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else {
            throw RecipesAPIError.invalidResponse
        }

        if !(200...299).contains(http.statusCode) {
            let message = String(data: data, encoding: .utf8) ?? "unknown"
            throw RecipesAPIError.serverError(status: http.statusCode, message: message)
        }
    }

    func logoutSession(refreshToken: String) async throws {
        var request = URLRequest(url: baseURL.appendingPathComponent("users/logout/session"))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(RefreshRequestPayload(refreshToken: refreshToken))

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else {
            throw RecipesAPIError.invalidResponse
        }

        if !(200...299).contains(http.statusCode) {
            let message = String(data: data, encoding: .utf8) ?? "unknown"
            throw RecipesAPIError.serverError(status: http.statusCode, message: message)
        }
    }

    func listSessions(userID: String, token: String) async throws -> [SessionItem] {
        var request = URLRequest(url: baseURL.appendingPathComponent("users/\(userID)/sessions"))
        request.httpMethod = "GET"
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        let (data, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else {
            throw RecipesAPIError.invalidResponse
        }

        if !(200...299).contains(http.statusCode) {
            let message = String(data: data, encoding: .utf8) ?? "unknown"
            throw RecipesAPIError.serverError(status: http.statusCode, message: message)
        }

        return try JSONDecoder().decode([SessionItem].self, from: data)
    }
}
