import Foundation

// MARK: - prayer-times next --json

struct NextPrayerResponse: Codable, Equatable {
    let prayer: String
    let time: String
    let remaining: String
}

// MARK: - prayer-times --json

struct TodayResponse: Codable, Equatable {
    let location: LocationInfo
    let date: DateInfo
    let timings: [String: String]
    let current: String?
    let next: NextPrayerInfo?
}

struct LocationInfo: Codable, Equatable {
    let city: String?
    let country: String?
    let timezone: String
    let latitude: Double
    let longitude: Double
}

struct DateInfo: Codable, Equatable {
    let gregorian: String
    let hijri: String
}

struct NextPrayerInfo: Codable, Equatable {
    let prayer: String
    let time: String
    let remaining: String
}

// MARK: - Prayer display helpers

/// A single prayer entry parsed from the today response, ordered chronologically.
struct PrayerEntry: Identifiable, Equatable {
    let id: String // prayer name
    let name: String
    let time: String

    /// The canonical display order for prayers.
    static let displayOrder = ["fajr", "sunrise", "dhuhr", "asr", "maghrib", "isha"]

    var sortIndex: Int {
        Self.displayOrder.firstIndex(of: id) ?? 99
    }
}

extension TodayResponse {
    /// Returns prayer entries sorted in chronological order.
    var sortedPrayers: [PrayerEntry] {
        timings.map { key, value in
            PrayerEntry(id: key, name: key.capitalized, time: value)
        }
        .sorted { $0.sortIndex < $1.sortIndex }
    }
}
