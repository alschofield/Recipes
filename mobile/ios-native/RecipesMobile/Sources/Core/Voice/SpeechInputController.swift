import AVFoundation
import Foundation
import Speech

@MainActor
final class SpeechInputController: NSObject, ObservableObject {
    enum SpeechInputError: LocalizedError {
        case unavailable
        case permissionDenied
        case recognizerInit

        var errorDescription: String? {
            switch self {
            case .unavailable:
                return "Speech recognition is unavailable on this device."
            case .permissionDenied:
                return "Microphone or speech permission denied."
            case .recognizerInit:
                return "Unable to initialize speech recognition."
            }
        }
    }

    @Published var isListening: Bool = false

    private let audioEngine = AVAudioEngine()
    private var recognitionRequest: SFSpeechAudioBufferRecognitionRequest?
    private var recognitionTask: SFSpeechRecognitionTask?
    private var recognizer: SFSpeechRecognizer?

    func stop() {
        recognitionTask?.cancel()
        recognitionTask = nil

        recognitionRequest?.endAudio()
        recognitionRequest = nil

        if audioEngine.isRunning {
            audioEngine.stop()
        }
        audioEngine.inputNode.removeTap(onBus: 0)
        isListening = false
    }

    func start(onTranscript: @escaping (String) -> Void) async throws {
        stop()
        VoiceTelemetryStore.incrementStart()

        let localeCandidates = [Locale.current, Locale(identifier: "en-US")]
        let selectedRecognizer = localeCandidates
            .compactMap { SFSpeechRecognizer(locale: $0) }
            .first { $0.isAvailable }

        guard let recognizer = selectedRecognizer else {
            VoiceTelemetryStore.incrementFailure()
            throw SpeechInputError.unavailable
        }
        self.recognizer = recognizer

        let hasPermissions = await requestPermissionsIfNeeded()
        guard hasPermissions else {
            VoiceTelemetryStore.incrementPermissionDenied()
            throw SpeechInputError.permissionDenied
        }

        let request = SFSpeechAudioBufferRecognitionRequest()
        request.shouldReportPartialResults = true
        recognitionRequest = request

        let inputNode = audioEngine.inputNode
        let format = inputNode.outputFormat(forBus: 0)
        inputNode.removeTap(onBus: 0)
        inputNode.installTap(onBus: 0, bufferSize: 1024, format: format) { [weak self] buffer, _ in
            self?.recognitionRequest?.append(buffer)
        }

        audioEngine.prepare()
        try audioEngine.start()
        isListening = true

        guard let recognitionRequest else {
            throw SpeechInputError.recognizerInit
        }

        var didMarkSuccess = false
        recognitionTask = recognizer.recognitionTask(with: recognitionRequest) { [weak self] result, error in
            guard let self else { return }

            if let result {
                Task { @MainActor in
                    onTranscript(result.bestTranscription.formattedString)
                }
                if result.isFinal {
                    if !didMarkSuccess {
                        VoiceTelemetryStore.incrementSuccess()
                        didMarkSuccess = true
                    }
                    self.stop()
                }
            }

            if error != nil {
                VoiceTelemetryStore.incrementFailure()
                self.stop()
            }
        }
    }

    private func requestPermissionsIfNeeded() async -> Bool {
        let speechStatus = await withCheckedContinuation { continuation in
            SFSpeechRecognizer.requestAuthorization { status in
                continuation.resume(returning: status)
            }
        }

        guard speechStatus == .authorized else {
            return false
        }

        let micGranted = await withCheckedContinuation { continuation in
            AVAudioSession.sharedInstance().requestRecordPermission { granted in
                continuation.resume(returning: granted)
            }
        }
        return micGranted
    }
}
