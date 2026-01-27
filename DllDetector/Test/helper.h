#pragma once
#include "CppUnitTest.h"
#include <string>
#include <filesystem>
#include <vector>


// Namespace for the Microsoft Unit Test Framework
using namespace Microsoft::VisualStudio::CppUnitTestFramework;
namespace fs = std::filesystem;


// Helper to get the absolute path to a test file using std::filesystem
fs::path GetTestFilePath(const std::wstring& fileName) {
    // Assumes the Sample folder is in the solution directory
    // This is more robust than manual string manipulation
    std::filesystem::path solutionDir = std::filesystem::current_path().parent_path().parent_path().parent_path().parent_path();
    return solutionDir / L"Test/Samples" / fileName;
}

// Helper to assert that a file exists for a test to be valid
void AssertTestFileExists(const fs::path& path) {
    std::wstring message = L"Test file not found: " + path.filename().wstring() + L". Please run the Setup-TestFiles.ps1 script.";
    Assert::IsTrue(fs::exists(path), message.c_str());
}

// Helper to make checking for anomalies in the results vector cleaner
bool ContainsAnomaly(const std::vector<PeAnalysis::Anomaly>& anomalies, PeAnalysis::Anomaly target) {
    return std::find(anomalies.begin(), anomalies.end(), target) != anomalies.end();
}
