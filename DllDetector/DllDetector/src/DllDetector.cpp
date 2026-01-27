/**
* @file DllDetector.cpp
* @brief Core implementation of the Resilient PE File Verifier library.
*/

#include "DllDetector/DllDetector.h"
#include "PeUtils.h"
#include <Windows.h>
#include <utility>

namespace PeAnalysis {
    // Main analysis function for wide-character paths
    Result IsFileDllW(const std::wstring& filePath, const Config& config) {
        Result result;

        // --- Phase 1: File & PE Structural Integrity Validation ---

        // Using a custom RAII wrapper for the file handle
        PeAnalysis::Internal::SafeFileHandle safeHandle(
            CreateFileW(filePath.c_str(), GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL)
        );

        if (!safeHandle) {
            result.verdict = Verdict::FILE_ERROR;
            return result;
        }

        LARGE_INTEGER fileSize;
        if (!GetFileSizeEx(safeHandle.get(), &fileSize) || fileSize.QuadPart < sizeof(IMAGE_DOS_HEADER)) {
            result.verdict = Verdict::NOT_A_PE;
            return result;
        }

        // Limit file size to avoid excessive memory allocation for huge files.
        const DWORD maxFileSize = config.maxFileSizeInMB * 1024 * 1024;
        if (fileSize.QuadPart > maxFileSize) {
            result.verdict = Verdict::FILE_ERROR;
            return result;
        }

        std::vector<char> fileBuffer(static_cast<size_t>(fileSize.QuadPart));
        DWORD bytesRead = 0;
        if (!ReadFile(safeHandle.get(), fileBuffer.data(), static_cast<DWORD>(fileBuffer.size()), &bytesRead, NULL) || bytesRead != fileBuffer.size()) {
            result.verdict = Verdict::FILE_ERROR;
            return result;
        }

        const auto* dosHeader = reinterpret_cast<const IMAGE_DOS_HEADER*>(fileBuffer.data());
         if (dosHeader->e_magic != IMAGE_DOS_SIGNATURE) {
            result.verdict = Verdict::NOT_A_PE;
            return result;
        }

        // Boundary check for the NT header offset
        if (dosHeader->e_lfanew <= 0 || static_cast<uint64_t>(dosHeader->e_lfanew) + sizeof(IMAGE_NT_HEADERS) > fileBuffer.size()) {
            result.verdict = Verdict::MALFORMED;
            return result;
        }

        const auto* ntHeader = reinterpret_cast<const IMAGE_NT_HEADERS*>(fileBuffer.data() + dosHeader->e_lfanew);
        if (ntHeader->Signature != IMAGE_NT_SIGNATURE) {
            result.verdict = Verdict::MALFORMED;
            return result;
        }

        if (ntHeader->FileHeader.NumberOfSections > 96) { // Windows loader limit
            result.verdict = Verdict::MALFORMED;
            return result;
        }

        // Boundary check for the section table itself
        DWORD sectionTableOffset = dosHeader->e_lfanew + sizeof(DWORD) + sizeof(IMAGE_FILE_HEADER) + ntHeader->FileHeader.SizeOfOptionalHeader;
        DWORD sectionTableSize = ntHeader->FileHeader.NumberOfSections * sizeof(IMAGE_SECTION_HEADER);
        if (sectionTableOffset + sectionTableSize > fileBuffer.size()) {
            result.verdict = Verdict::MALFORMED;
            return result;
        }

        // --- Phase 2: Corroboration of Primary DLL Indicators ---

        bool has_dll_flag = (ntHeader->FileHeader.Characteristics & IMAGE_FILE_DLL) != 0;
        bool is_entry_point_zero = ntHeader->OptionalHeader.AddressOfEntryPoint == 0;
        bool has_exports = false;

        const IMAGE_DATA_DIRECTORY& exportDir = ntHeader->OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT];
        if (exportDir.VirtualAddress != 0 && exportDir.Size != 0) {
            DWORD exportDirOffset = PeAnalysis::Internal::RvaToOffset(ntHeader, fileBuffer, exportDir.VirtualAddress);
            if (exportDirOffset != 0 && exportDirOffset + sizeof(IMAGE_EXPORT_DIRECTORY) <= fileBuffer.size()) {
                const auto* exportDesc = reinterpret_cast<const IMAGE_EXPORT_DIRECTORY*>(fileBuffer.data() + exportDirOffset);
                if (exportDesc->NumberOfFunctions > 0 || exportDesc->NumberOfNames > 0) {
                    has_exports = true;
                }
            }
        }

        // --- Phase 3: Anomaly and Deception Detection ---
        if (config.options != PeAnalysis::Options::NONE) {
            if (!has_dll_flag && has_exports) {
                result.anomalies.push_back(Anomaly::HEADER_CONTRADICTION_DLL_FLAG_VS_EXPORTS);
            }
            if (!has_dll_flag && is_entry_point_zero && !has_exports) {
                result.anomalies.push_back(Anomaly::HEADER_CONTRADICTION_DLL_FLAG_VS_ENTRYPOINT);
            }

            const IMAGE_SECTION_HEADER* sectionHeader = IMAGE_FIRST_SECTION(ntHeader);
            const IMAGE_SECTION_HEADER* entryPointSection = nullptr;
            DWORD entryPointRva = ntHeader->OptionalHeader.AddressOfEntryPoint;

            // --- Entry Point and Section Analysis Loop ---
            for (WORD i = 0; i < ntHeader->FileHeader.NumberOfSections; ++i) {
                const auto& currentSection = sectionHeader[i];
                DWORD characteristics = currentSection.Characteristics;

                // --- 1. Entry Point Location Check ---
                if (entryPointRva >= currentSection.VirtualAddress && entryPointRva < (currentSection.VirtualAddress + currentSection.Misc.VirtualSize)) {
                    entryPointSection = &currentSection;
                    // Check if the entry point is in a non-code section (highly suspicious)
                    if (strcmp(reinterpret_cast<const char*>(currentSection.Name), ".text") != 0 &&
                        strcmp(reinterpret_cast<const char*>(currentSection.Name), "CODE") != 0) {
                        if (std::find(result.anomalies.begin(), result.anomalies.end(), Anomaly::SUSPICIOUS_ENTRY_POINT) == result.anomalies.end()) {
                            result.anomalies.push_back(Anomaly::SUSPICIOUS_ENTRY_POINT);
                            result.is_likely_packed = true;
                        }
                    }
                }

                // A legitimate BSS section is uninitialized AND should NOT be executable.
                if ((characteristics & IMAGE_SCN_CNT_UNINITIALIZED_DATA) && !(characteristics & IMAGE_SCN_MEM_EXECUTE)) {
                    continue; // Skip further anomaly checks for this benign section.
                }

                // --- 2. Zero-Size Section Anomaly (Replaces the multiplier) ---
                if (currentSection.SizeOfRawData == 0 && currentSection.Misc.VirtualSize > 1024) {
                    result.is_likely_packed = true;
                    if (std::find(result.anomalies.begin(), result.anomalies.end(), Anomaly::PACKED_SECTION_SIZE) == result.anomalies.end()) {
                        result.anomalies.push_back(Anomaly::PACKED_SECTION_SIZE);
                    }
                }

                // --- 3. Entropy Check (Now linked to the Entry Point Section) ---
                if ((config.options & Options::ENTROPY_CHECK) != Options::NONE && currentSection.SizeOfRawData > 0) {
                    double entropy = PeAnalysis::Internal::CalculateEntropy(
                        reinterpret_cast<const unsigned char*>(fileBuffer.data() + currentSection.PointerToRawData),
                        currentSection.SizeOfRawData
                    );

                    // High entropy in an executable section is a strong indicator of packing.
                    if (entropy > 7.0 && (characteristics & IMAGE_SCN_MEM_EXECUTE)) {
                        result.is_likely_packed = true;
                        if (std::find(result.anomalies.begin(), result.anomalies.end(), Anomaly::HIGH_ENTROPY_SECTION) == result.anomalies.end()) {
                            result.anomalies.push_back(Anomaly::HIGH_ENTROPY_SECTION);
                        }
                    }
                }

                // --- 4. RWX Section Check ---
                if ((config.options & Options::RWX_SECTION_CHECK) != Options::NONE) {
                    if ((characteristics & IMAGE_SCN_MEM_READ) &&
                        (characteristics & IMAGE_SCN_MEM_WRITE) &&
                        (characteristics & IMAGE_SCN_MEM_EXECUTE)) {
                        if (std::find(result.anomalies.begin(), result.anomalies.end(), Anomaly::RWX_SECTION) == result.anomalies.end()) {
                            result.anomalies.push_back(Anomaly::RWX_SECTION);
                        }
                    }
                }
            }
        }

        // ==============================================================================
    // --- Phase 4: The Final Verdict ---
    // This logic replaces the flawed weighted scoring model.
    // It prioritizes the definitive IMAGE_FILE_DLL flag and treats
    // contradictions as anomalies rather than verdict-changers.
    // ==============================================================================

        if (has_dll_flag) {
            // Rule 1: If the DLL flag is set, it's a DLL.
            result.verdict = Verdict::IS_DLL;

            // Check for contradictions for a DLL.
            // A DLL that is not an EXE and has no exports and a zero entry point might be a resource-only DLL.
            if (!has_exports && is_entry_point_zero) {
                if (std::find(result.anomalies.begin(), result.anomalies.end(), Anomaly::HEADER_CONTRADICTION_DLL_FLAG_VS_ENTRYPOINT) == result.anomalies.end()) {
                    result.anomalies.push_back(Anomaly::HEADER_CONTRADICTION_DLL_FLAG_VS_ENTRYPOINT);
                }
            }
        }
        else {
            // Rule 2: If the DLL flag is NOT set, it's an EXE.
            result.verdict = Verdict::IS_EXE;

            // Check for contradictions for an EXE.
            // An EXE with an export table is unusual but valid. We flag it as an anomaly.
            if (has_exports) {
                if (std::find(
                    result.anomalies.begin(),
                    result.anomalies.end(),
                    Anomaly::HEADER_CONTRADICTION_EXE_WITH_EXPORTS) == result.anomalies.end())
                {
                    result.anomalies.push_back(Anomaly::HEADER_CONTRADICTION_EXE_WITH_EXPORTS);
                }
            }
        }

        // The confidence score logic is desired for informational purposes
        int score = 0;
        if (has_dll_flag) score += 10; // High confidence from the flag itself
        if (has_exports) score += 5;
        if (is_entry_point_zero) score += 2;
        result.confidence_score = score;

        // Override verdict if high-confidence packing/security anomalies were found
        if (config.options != Options::NONE) {
            if (result.is_likely_packed || !result.anomalies.empty()) {
                result.verdict = Verdict::MALFORMED;
            }
        }

        return result;
    }

    // Wrapper function for multi-byte/UTF-8 paths
    Result IsFileDllA(const std::string& filePath, const Config& config) {
        Result result;
        if (filePath.empty()) {
            result.verdict = Verdict::FILE_ERROR;
            return result;
        }

        // First call to get the required buffer size for the wide string
        int wideCharCount = MultiByteToWideChar(CP_UTF8, MB_ERR_INVALID_CHARS, filePath.c_str(), -1, NULL, 0);
        if (wideCharCount == 0) {
            result.verdict = Verdict::MALFORMED; // Indicates an invalid path string
            return result;
        }

        std::wstring wFilePath(wideCharCount, L'\0');

        // Second call to perform the actual conversion
        if (MultiByteToWideChar(CP_UTF8, MB_ERR_INVALID_CHARS, filePath.c_str(), -1, wFilePath.data(), wideCharCount) == 0) {
            result.verdict = Verdict::MALFORMED;
            return result;
        }

        // Remove the extra null terminator from the wstring's internal count
        wFilePath.resize(wcslen(wFilePath.c_str()));

        return IsFileDllW(wFilePath, config);
    }


}