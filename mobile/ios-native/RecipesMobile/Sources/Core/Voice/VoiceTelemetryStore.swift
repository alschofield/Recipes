import Foundation

enum VoiceTelemetryStore {
    private static let startKey = "voice.stt.start"
    private static let successKey = "voice.stt.success"
    private static let failureKey = "voice.stt.failure"
    private static let permissionDeniedKey = "voice.stt.permission_denied"

    static func incrementStart() {
        increment(startKey)
    }

    static func incrementSuccess() {
        increment(successKey)
    }

    static func incrementFailure() {
        increment(failureKey)
    }

    static func incrementPermissionDenied() {
        increment(permissionDeniedKey)
    }

    private static func increment(_ key: String) {
        let next = UserDefaults.standard.integer(forKey: key) + 1
        UserDefaults.standard.set(next, forKey: key)
    }
}
