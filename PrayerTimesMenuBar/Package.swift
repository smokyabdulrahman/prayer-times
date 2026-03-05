// swift-tools-version: 5.10

import PackageDescription

let package = Package(
    name: "PrayerTimesMenuBar",
    platforms: [
        .macOS(.v14),
    ],
    targets: [
        .executableTarget(
            name: "PrayerTimesMenuBar",
            path: "Sources",
            resources: [
                .copy("../Resources/Info.plist"),
            ]
        ),
    ]
)
