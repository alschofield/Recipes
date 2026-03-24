import SwiftUI
import UIKit

struct SearchView: View {
    @State private var ingredientsRaw: String = "chicken, rice, garlic"
    @State private var mode: String = "strict"
    @State private var complex: Bool = false
    @State private var loading: Bool = false
    @State private var voiceError: String?
    @State private var error: String?
    @State private var payload: SearchResponse?
    @StateObject private var speechInput = SpeechInputController()

    private let apiClient = RecipesAPIClient()

    var body: some View {
        NavigationStack {
            List {
                Section("Ingredients") {
                    TextField("chicken, rice, garlic", text: $ingredientsRaw)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled(true)
                        .accessibilityLabel("Ingredients input")

                    Button(speechInput.isListening ? "Stop dictation" : "Dictate ingredients") {
                        Task { await toggleVoiceInput() }
                    }
                    .disabled(loading)
                    .accessibilityLabel("Dictate ingredients")
                    .accessibilityHint("Use speech to fill ingredients input")
                }

                Section("Mode") {
                    Picker("Mode", selection: $mode) {
                        Text("Strict").tag("strict")
                        Text("Inclusive").tag("inclusive")
                    }
                    .pickerStyle(.segmented)
                    .accessibilityLabel("Search mode")

                    Toggle("Complex", isOn: $complex)
                        .accessibilityLabel("Complex mode")
                }

                Section {
                    Button(loading ? "Finding matches..." : "Find matches") {
                        Task { await runSearch() }
                    }
                    .disabled(loading)
                    .accessibilityLabel("Run recipe search")
                }

                if let error {
                    Section("Error") {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }

                if let voiceError {
                    Section("Voice") {
                        Text(voiceError)
                            .foregroundStyle(.red)
                    }
                }

                if let payload {
                    Section("Results") {
                        Text("Mode: \(payload.mode)")
                        Text("Total: \(payload.pagination.total)")
                    }

                    Section("Top Matches") {
                        ForEach(payload.results.prefix(20)) { item in
                            VStack(alignment: .leading, spacing: 4) {
                                Text(item.name)
                                    .font(.headline)
                                let percent = Int((item.matchPercent ?? 0) * 100)
                                Text("\(item.source) • \(percent)% match")
                                    .font(.footnote)
                                    .foregroundStyle(.secondary)
                            }
                            .padding(.vertical, 4)
                        }
                    }
                }
            }
            .listStyle(.insetGrouped)
            .scrollContentBackground(.hidden)
            .navigationTitle("Discover Recipes")
            .onChange(of: voiceError) { _, next in
                guard let next, !next.isEmpty else { return }
                UIAccessibility.post(notification: .announcement, argument: "Voice error. \(next)")
            }
            .onChange(of: payload?.pagination.total) { _, total in
                guard let total else { return }
                UIAccessibility.post(notification: .announcement, argument: "Search completed. \(total) results.")
            }
            .onDisappear {
                speechInput.stop()
            }
        }
    }

    private func runSearch() async {
        loading = true
        defer { loading = false }
        error = nil

        let ingredients = ingredientsRaw
            .split(separator: ",")
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines).lowercased() }
            .filter { !$0.isEmpty }

        do {
            payload = try await apiClient.search(ingredients: ingredients, mode: mode, complex: complex)
        } catch {
            self.error = error.localizedDescription
        }
    }

    private func toggleVoiceInput() async {
        if speechInput.isListening {
            speechInput.stop()
            return
        }

        voiceError = nil
        do {
            try await speechInput.start { transcript in
                let merged = mergeIngredients(existing: ingredientsRaw, transcript: transcript)
                ingredientsRaw = merged
            }
        } catch {
            if let speechError = error as? SpeechInputController.SpeechInputError {
                switch speechError {
                case .unavailable:
                    voiceError = "Speech recognition unavailable on this device. Use manual typing for ingredients."
                case .permissionDenied:
                    voiceError = "Microphone or speech permission denied. You can continue with typed ingredients."
                case .recognizerInit:
                    voiceError = "Unable to initialize speech recognition. Please try again."
                }
            } else {
                voiceError = error.localizedDescription
            }
        }
    }
}

private func mergeIngredients(existing: String, transcript: String) -> String {
    let normalized = transcript
        .replacingOccurrences(of: " and ", with: ", ")
        .replacingOccurrences(of: ";", with: ",")
        .trimmingCharacters(in: .whitespacesAndNewlines)

    guard !normalized.isEmpty else {
        return existing
    }

    if existing.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
        return normalized
    }

    let suffix = existing.trimmingCharacters(in: .whitespacesAndNewlines).hasSuffix(",") ? " " : ", "
    return existing.trimmingCharacters(in: .whitespacesAndNewlines) + suffix + normalized
}
