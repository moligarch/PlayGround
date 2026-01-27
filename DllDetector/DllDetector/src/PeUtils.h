/**
 * @file PeUtils.h
 * @brief Internal utilities for PE file parsing.
 */

#pragma once

#include <Windows.h>
#include <vector>
#include <string>
#include <cstdint>

namespace PeAnalysis {
    namespace Internal {

        /**
         * @brief RAII wrapper for a Windows HANDLE to ensure it's always closed.
         */
        class SafeFileHandle {
        public:
            explicit SafeFileHandle(HANDLE handle);
            ~SafeFileHandle();

            // Disable copy and assignment
            SafeFileHandle(const SafeFileHandle&) = delete;
            SafeFileHandle& operator=(const SafeFileHandle&) = delete;

            HANDLE get() const;
            explicit operator bool() const;

        private:
            HANDLE m_handle;
        };

        /**
         * @brief Converts a Relative Virtual Address (RVA) to a raw file offset.
         * @param ntHeader Pointer to the PE file's NT Headers.
         * @param fileBuffer A vector containing the entire file's content.
         * @param rva The RVA to convert.
         * @return The raw file offset corresponding to the RVA, or 0 if invalid.
         */
        DWORD RvaToOffset(const IMAGE_NT_HEADERS* ntHeader, const std::vector<char>& fileBuffer, DWORD rva);

        /**
         * @brief Calculates the Shannon entropy for a block of data.
         * @param data Pointer to the data buffer.
         * @param size The size of the data buffer.
         * @return The calculated entropy value (0.0 to 8.0).
         */
        double CalculateEntropy(const unsigned char* data, size_t size);

    } // namespace Internal
} // namespace PeAnalysis