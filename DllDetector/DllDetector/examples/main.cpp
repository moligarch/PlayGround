/**
 * @file main.cpp
 * @brief Example usage of the DllDetector library.
 *
 * This program takes a single command-line argument (a file path) and
 * uses the DllDetector library to analyze the file, printing the results
 * to the console.
 */

#include <iostream>
#include <string>
#include <filesystem>
#include <map>
#include <Windows.h>
#include "DllDetector/DllDetector.h"

namespace fs = std::filesystem;
using namespace PeAnalysis;

/**
 * @brief Prints the command-line usage instructions for the program.
 * @param appName The name of the executable (from argv[0]).
 */
void PrintUsage(const char* appName) {
    std::cerr << "DLL Verifier Tool\n\n";
    std::cerr << "Usage: " << fs::path(appName).filename().string() << " [options] <file_or_directory_path>\n\n";
    std::cerr << "Description:\n";
    std::cerr << "  Analyzes a PE file or recursively scans a directory to verify if files are DLLs.\n\n";
    std::cerr << "Options:\n";
    std::cerr << "  --entropy-check    Enable entropy calculation to detect packed sections. (CPU intensive)\n";
    std::cerr << "  --rwx-check        Enable check for suspicious Read-Write-Execute sections.\n";
    std::cerr << "  --check-all        Enable all available checks. (This is the default behavior)\n";
    std::cerr << "  --file-size <MB>   Set the max file size in Megabytes for analysis. Default is 100.\n";
    std::cerr << "  --help             Display this help message.\n\n";
    std::cerr << "Example:\n";
    std::cerr << "  " << fs::path(appName).filename().string() << " --entropy-check C:\\Windows\\System32\n";
}

/**
 * @brief Processes a single file, calls the verifier, and prints the formatted result.
 * @param filePath The path to the file to process.
 * @param config The analysis configuration.
 * @param counters A map to track statistics for the final summary.
 */
void ProcessFile(const fs::path& filePath, const Config& config, std::map<Verdict, int>& counters) {
    std::cout << "-> Checking: " << filePath.string() << "\n";

    // Use IsFileDllA as we are getting paths from the filesystem API which can provide std::string
    Result result = IsFileDllW(filePath.wstring(), config);

    counters[result.verdict]++;

    HANDLE hConsole = GetStdHandle(STD_OUTPUT_HANDLE); // Get console handle

    // Print verdict with color coding for clarity
    std::cout << "   Verdict: ";

    // Set color based on verdict
    WORD color = 7; // Default: light gray
    switch (result.verdict) {
    case Verdict::IS_DLL:       color = 10; break; // Green
    case Verdict::IS_EXE:       color = 11; break; // Cyan
    case Verdict::MALFORMED:    color = 12; break; // Red
    case Verdict::UNKNOWN_PE:   color = 14; break; // Yellow
    }
    SetConsoleTextAttribute(hConsole, color); // Apply the color

    // Print the text
    switch (result.verdict) {
    case Verdict::IS_DLL:       std::cout << "IS_DLL"; break;
    case Verdict::IS_EXE:       std::cout << "IS_EXE"; break;
    case Verdict::MALFORMED:    std::cout << "MALFORMED"; break;
    case Verdict::UNKNOWN_PE:   std::cout << "UNKNOWN_PE"; break;
    case Verdict::NOT_A_PE:     std::cout << "NOT_A_PE"; break;
    default: break;
    }

    SetConsoleTextAttribute(hConsole, 7); // Reset to default color

    /*if (result.verdict != Verdict::NOT_A_PE && result.verdict != Verdict::ACCESS_DENIED && result.verdict != Verdict::READ_ERROR) {
        std::cout << " (Confidence: " << result.confidence_score << ")\n";
    }
    else {
        std::cout << "\n";
    }*/

    if (result.is_likely_packed) {
        std::cout << "   Note: File is likely packed.\n";
    }

    if (!result.anomalies.empty()) {
        std::cout << "   Anomalies Detected:\n";
        for (const auto& anomaly : result.anomalies) {
            std::cout << "     - ";
            switch (anomaly) {
            case Anomaly::PACKED_SECTION_SIZE:  std::cout << "Packed Section (Size Discrepancy)"; break;
            case Anomaly::HIGH_ENTROPY_SECTION: std::cout << "High Entropy Section"; break;
            case Anomaly::RWX_SECTION:           std::cout << "Read-Write-Execute (RWX) Section"; break;
            case Anomaly::HEADER_CONTRADICTION_DLL_FLAG_VS_EXPORTS:    std::cout << "Header Contradiction (No DLL flag but has exports)"; break;
            case Anomaly::HEADER_CONTRADICTION_DLL_FLAG_VS_ENTRYPOINT: std::cout << "Header Contradiction (No DLL flag and zero entry point)"; break;
            }
            std::cout << "\n";
        }
    }
    std::cout << "\n";
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        PrintUsage(argv[0]);
        return 1;
    }

    Config config;
    config.options = Options::NONE; // Start with no options, add them as they are parsed
    std::string targetPathStr;
    bool optionsManuallySet = false;

    // --- Argument Parsing ---
    try {
        for (int i = 1; i < argc; ++i) {
            std::string arg = argv[i];
            if (arg == "--help") {
                PrintUsage(argv[0]);
                return 0;
            }
            else if (arg == "--entropy-check") {
                config.options |= Options::ENTROPY_CHECK;
                optionsManuallySet = true;
            }
            else if (arg == "--rwx-check") {
                config.options |= Options::RWX_SECTION_CHECK;
                optionsManuallySet = true;
            }
            else if (arg == "--check-all") {
                config.options = Options::ALL;
                optionsManuallySet = true;
            }
            else if (arg == "--file-size") {
                if (i + 1 < argc) {
                    config.maxFileSizeInMB = std::stoul(argv[++i]);
                }
                else {
                    throw std::runtime_error("--file-size requires a value.");
                }
            }
            else if (arg[0] == '-') {
                throw std::runtime_error("Unknown option: " + arg);
            }
            else {
                // Assume it's the path
                if (!targetPathStr.empty()) {
                    throw std::runtime_error("Target path specified more than once.");
                }
                targetPathStr = arg;
            }
        }
    }
    catch (const std::exception& e) {
        std::cerr << "Error parsing arguments: " << e.what() << "\n\n";
        PrintUsage(argv[0]);
        return 1;
    }

    if (targetPathStr.empty()) {
        std::cerr << "Error: No file or directory path provided.\n\n";
        PrintUsage(argv[0]);
        return 1;
    }

    // --- Path Processing ---
    fs::path targetPath(targetPathStr);
    auto opt = fs::directory_options::skip_permission_denied;
    std::error_code ec;
    std::map<Verdict, int> counters;
    int filesScanned = 0;

    if (!fs::exists(targetPath, ec)) {
        std::cerr << "Error: Path does not exist: " << targetPathStr << "\n";
        return 1;
    }

    if (fs::is_directory(targetPath, ec)) {
        std::cout << "Scanning directory recursively: " << targetPath.string() << "\n\n";
        try {
            for (const auto& entry : fs::recursive_directory_iterator(targetPath, opt, ec)) {
                if (entry.is_regular_file()) {
                    ProcessFile(entry.path(), config, counters);
                    filesScanned++;
                }
            }
        }
        catch (const std::exception& e) {
            std::cerr << "An error occurred during directory traversal: " << e.what() << "\n";
        }
    }
    else if (fs::is_regular_file(targetPath, ec)) {
        ProcessFile(targetPath, config, counters);
        filesScanned = 1;
    }
    else {
        std::cerr << "Error: Path is not a regular file or directory: " << targetPathStr << "\n";
        return 1;
    }

    // --- Final Summary ---
    std::cout << "========================================\n";
    std::cout << "           ANALYSIS SUMMARY\n";
    std::cout << "========================================\n";
    std::cout << "Total Files Scanned: " << filesScanned << "\n";
    if (counters[Verdict::IS_DLL] > 0)       std::cout << "DLLs Found:          " << counters[Verdict::IS_DLL] << "\n";
    if (counters[Verdict::IS_EXE] > 0)       std::cout << "EXEs Found:          " << counters[Verdict::IS_EXE] << "\n";
    if (counters[Verdict::MALFORMED] > 0)    std::cout << "Malformed PEs:       " << counters[Verdict::MALFORMED] << "\n";
    if (counters[Verdict::UNKNOWN_PE] > 0)   std::cout << "Unknown PEs:         " << counters[Verdict::UNKNOWN_PE] << "\n";
    std::cout << "========================================\n";

    return 0;
}