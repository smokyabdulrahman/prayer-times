import Foundation

/// Errors that can occur when interacting with the prayer-times CLI.
enum CLIError: LocalizedError {
    case binaryNotFound
    case executionFailed(String)
    case decodingFailed(String)

    var errorDescription: String? {
        switch self {
        case .binaryNotFound:
            "prayer-times binary not found. Ensure it is bundled in Resources or available on PATH."
        case .executionFailed(let message):
            "CLI execution failed: \(message)"
        case .decodingFailed(let message):
            "Failed to decode CLI output: \(message)"
        }
    }
}

/// Service that shells out to the prayer-times CLI binary and parses JSON output.
@Observable
final class CLIService {
    /// Cached path to the CLI binary, resolved once on first use.
    private var resolvedBinaryPath: String?

    /// Fetches the next prayer using `prayer-times next --json`.
    func fetchNext() async throws -> NextPrayerResponse {
        let output = try await run(arguments: ["next", "--json"])
        return try decode(NextPrayerResponse.self, from: output)
    }

    /// Fetches the next prayer formatted as a display string.
    /// Uses `prayer-times next --format <format>` (plain text, not JSON).
    func fetchNextFormatted(format: String) async throws -> String {
        let output = try await run(arguments: ["next", "--format", format])
        return output.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    /// Fetches today's full prayer schedule using `prayer-times --json`.
    func fetchToday() async throws -> TodayResponse {
        let output = try await run(arguments: ["--json"])
        return try decode(TodayResponse.self, from: output)
    }

    // MARK: - Private

    /// Locates the prayer-times binary.
    /// Priority: app bundle Resources > common install paths > PATH lookup.
    private func findBinary() throws -> String {
        if let cached = resolvedBinaryPath {
            return cached
        }

        let candidates: [String] = [
            // Bundled inside .app/Contents/Resources/
            Bundle.main.resourcePath.map { $0 + "/prayer-times" },
            // Common install locations
            "/usr/local/bin/prayer-times",
            "/opt/homebrew/bin/prayer-times",
            // Go default install path
            NSHomeDirectory() + "/go/bin/prayer-times",
        ].compactMap { $0 }

        for path in candidates {
            if FileManager.default.isExecutableFile(atPath: path) {
                resolvedBinaryPath = path
                return path
            }
        }

        // Try PATH via `which`
        if let whichPath = try? shellWhich("prayer-times") {
            resolvedBinaryPath = whichPath
            return whichPath
        }

        throw CLIError.binaryNotFound
    }

    /// Runs the CLI binary with the given arguments and returns stdout as a string.
    private func run(arguments: [String]) async throws -> String {
        let binaryPath = try findBinary()

        return try await withCheckedThrowingContinuation { continuation in
            DispatchQueue.global(qos: .userInitiated).async {
                do {
                    let process = Process()
                    process.executableURL = URL(fileURLWithPath: binaryPath)
                    process.arguments = arguments

                    let stdout = Pipe()
                    let stderr = Pipe()
                    process.standardOutput = stdout
                    process.standardError = stderr

                    try process.run()
                    process.waitUntilExit()

                    let outputData = stdout.fileHandleForReading.readDataToEndOfFile()
                    let errorData = stderr.fileHandleForReading.readDataToEndOfFile()

                    guard process.terminationStatus == 0 else {
                        let errorMessage = String(data: errorData, encoding: .utf8) ?? "Unknown error"
                        continuation.resume(throwing: CLIError.executionFailed(errorMessage.trimmingCharacters(in: .whitespacesAndNewlines)))
                        return
                    }

                    guard let output = String(data: outputData, encoding: .utf8) else {
                        continuation.resume(throwing: CLIError.decodingFailed("Unable to read stdout as UTF-8"))
                        return
                    }

                    continuation.resume(returning: output)
                } catch {
                    continuation.resume(throwing: error)
                }
            }
        }
    }

    /// Decodes a JSON string into the specified Codable type.
    private func decode<T: Decodable>(_ type: T.Type, from jsonString: String) throws -> T {
        guard let data = jsonString.data(using: .utf8) else {
            throw CLIError.decodingFailed("Invalid UTF-8 in JSON output")
        }
        do {
            return try JSONDecoder().decode(T.self, from: data)
        } catch {
            throw CLIError.decodingFailed(error.localizedDescription)
        }
    }

    /// Uses `which` to find a binary on PATH.
    private func shellWhich(_ name: String) throws -> String? {
        let process = Process()
        process.executableURL = URL(fileURLWithPath: "/usr/bin/which")
        process.arguments = [name]

        let pipe = Pipe()
        process.standardOutput = pipe
        process.standardError = FileHandle.nullDevice

        try process.run()
        process.waitUntilExit()

        guard process.terminationStatus == 0 else { return nil }

        let data = pipe.fileHandleForReading.readDataToEndOfFile()
        return String(data: data, encoding: .utf8)?.trimmingCharacters(in: .whitespacesAndNewlines)
    }
}
