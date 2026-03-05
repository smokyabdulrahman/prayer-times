import Foundation

/// Display format options for the menubar text.
/// Each case maps to a prayer-times CLI `--format` value.
enum DisplayFormat: String, CaseIterable, Identifiable {
    case nameAndTime = "name-and-time"
    case nameAndRemaining = "name-and-remaining"
    case shortNameAndTime = "short-name-and-time"
    case shortNameAndRemaining = "short-name-and-remaining"
    case timeRemaining = "time-remaining"
    case nextPrayerTime = "next-prayer-time"
    case full = "full"

    var id: String { rawValue }

    /// Human-readable label for the settings UI.
    var displayName: String {
        switch self {
        case .nameAndTime: "Name & Time"
        case .nameAndRemaining: "Name & Remaining"
        case .shortNameAndTime: "Short Name & Time"
        case .shortNameAndRemaining: "Short Name & Remaining"
        case .timeRemaining: "Time Remaining"
        case .nextPrayerTime: "Next Prayer Time"
        case .full: "Full"
        }
    }

    /// Example preview text for the settings UI.
    var preview: String {
        switch self {
        case .nameAndTime: "Asr 3:12 PM"
        case .nameAndRemaining: "Asr 2h 15m"
        case .shortNameAndTime: "A 3:12 PM"
        case .shortNameAndRemaining: "A 2h 15m"
        case .timeRemaining: "2h 15m"
        case .nextPrayerTime: "3:12 PM"
        case .full: "Asr 3:12 PM (2h 15m)"
        }
    }
}
