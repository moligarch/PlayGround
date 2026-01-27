/**
 * @file DllDetector.h
 * @brief Public API for the Resilient DLL File Verifier library.
 *
 * This header defines the core data structures and functions for verifying
 * whether a given file is a Dynamic-Link Library (DLL) or a standard
 * Executable (EXE) by performing a deep, resilient analysis of its
 * Portable Executable (PE) structure.
 */

#pragma once

#include <vector>
#include <string>
#include <cstdint>


namespace PeAnalysis {

    /**
     * @enum Options
     * @brief Defines optional, potentially expensive checks that can be enabled or disabled.
     * @note Use bitwise OR to combine options, e.g., Options::ENTROPY_CHECK | Options::RWX_SECTION_CHECK.
     */
    enum class Options : uint32_t {
        NONE                = 0,
        ENTROPY_CHECK       = (1 << 0), ///< Performs entropy calculation on sections to detect packing. (CPU intensive)
        RWX_SECTION_CHECK   = (1 << 1), ///< Checks for sections with Read, Write, and Execute permissions.
        ALL                 = ENTROPY_CHECK | RWX_SECTION_CHECK
    };

    // Enable bitwise operations for the strongly-typed Options enum
    inline Options operator|(Options a, Options b) { return static_cast<Options>(static_cast<uint32_t>(a) | static_cast<uint32_t>(b)); }
    inline Options operator&(Options a, Options b) { return static_cast<Options>(static_cast<uint32_t>(a) & static_cast<uint32_t>(b)); }
    inline Options& operator|=(Options& a, Options b) { a = a | b; return a; }

    /**
     * @struct Config
     * @brief Holds configuration settings for the PE file analysis.
     */
    struct Config {
        Options options = Options::NONE;
        uint32_t maxFileSizeInMB = 20;
    };

    /**
        * @enum Verdict
        * @brief Defines the final conclusion of the PE file analysis.
        */
    enum class Verdict {
        IS_DLL,         ///< The file is identified as a Dynamic-Link Library.
        IS_EXE,         ///< The file is identified as a standard Executable.
        UNKNOWN_PE,     ///< The file is a valid PE but its type is ambiguous.
        NOT_A_PE,       ///< The file is not a valid Portable Executable.
        MALFORMED,      ///< The file is a PE but contains structural errors or inconsistencies.
        FILE_ERROR,     ///< The file could not be opened/read.
    };

    /**
        * @enum Anomaly
        * @brief Defines specific anomalies or suspicious characteristics detected during parsing.
        *
        * These flags provide additional context about the file, which can be used
        * to assess its risk level, even if the primary verdict is not malicious.
        */
    enum class Anomaly {
        PACKED_SECTION_SIZE,                            ///< A section has a virtual size much larger than its raw size.
        HIGH_ENTROPY_SECTION,                           ///< A section has high Shannon entropy, suggesting encryption or compression.
        RWX_SECTION,                                    ///< A section is marked as Readable, Writable, and Executable.
        SUSPICIOUS_ENTRY_POINT,                         ///< The entry point is in a non-standard or suspicious section.
        HEADER_CONTRADICTION_DLL_FLAG_VS_EXPORTS,       ///< File lacks the DLL flag but has an export table.
        HEADER_CONTRADICTION_DLL_FLAG_VS_ENTRYPOINT,    ///< File lacks the DLL flag and has a zero entry point.
        HEADER_CONTRADICTION_EXE_WITH_EXPORTS           ///< An EXE with an export table was detected.
    };

    /**
        * @struct Result
        * @brief Encapsulates the complete result of a PE file verification.
        *
        * This structure provides a comprehensive summary of the analysis, including
        * the final verdict, a confidence score, and a list of any detected anomalies.
        */
    struct Result {
        Verdict verdict = Verdict::UNKNOWN_PE; ///< The final verdict of the analysis.
        int confidence_score = 0;              ///< A weighted score indicating confidence in the verdict.
        bool is_likely_packed = false;         ///< True if packing indicators were found.
        std::vector<Anomaly> anomalies;        ///< A list of detected suspicious characteristics.
    };

    /**
     * @brief Verifies if a file is a DLL using deep PE structure analysis.
     * @param filePath The full path to the file to be checked (Unicode, wide-character string).
     * @param config Configuration settings to control the analysis flow.
     * @return A Result structure containing the verdict and detailed analysis findings.
     */
    Result IsFileDllW(const std::wstring& filePath, const Config& config = Config{});

    /**
     * @brief Verifies if a file is a DLL using deep PE structure analysis (multi-byte/UTF-8 path).
     * @param filePath The full path to the file to be checked (assumed to be UTF-8).
     * @param config Configuration settings to control the analysis flow.
     * @return A Result structure containing the verdict and detailed analysis findings.
     */
    Result IsFileDllA(const std::string& filePath, const Config& config = Config{});

}