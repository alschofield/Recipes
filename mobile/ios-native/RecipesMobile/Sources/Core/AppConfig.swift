import Foundation

enum AppConfig {
    static var apiBaseURL: URL {
        let value = Bundle.main.object(forInfoDictionaryKey: "RECIPES_API_BASE_URL") as? String
        let trimmed = value?.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        return URL(string: trimmed.isEmpty ? "https://api.yourdomain.com" : trimmed)!
    }
}
