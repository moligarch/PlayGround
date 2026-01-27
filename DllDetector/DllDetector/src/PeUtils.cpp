/**
 * @file PeUtils.cpp
 * @brief Implementation of internal PE parsing utilities.
 */

#include "PeUtils.h"
#include <cmath>
#include <numeric>
#include <algorithm>

namespace PeAnalysis {
    namespace Internal {

        // SafeFileHandle Implementation
        SafeFileHandle::SafeFileHandle(HANDLE handle) : m_handle(handle) {}

        SafeFileHandle::~SafeFileHandle() {
            if (m_handle != INVALID_HANDLE_VALUE) {
                CloseHandle(m_handle);
            }
        }

        HANDLE SafeFileHandle::get() const { return m_handle; }

        SafeFileHandle::operator bool() const { return m_handle != INVALID_HANDLE_VALUE; }


        // RvaToOffset Implementation
        DWORD RvaToOffset(const IMAGE_NT_HEADERS* ntHeader, const std::vector<char>& fileBuffer, DWORD rva) {
            if (!ntHeader) {
                return 0;
            }

            const IMAGE_SECTION_HEADER* sectionHeader = IMAGE_FIRST_SECTION(ntHeader);
            for (WORD i = 0; i < ntHeader->FileHeader.NumberOfSections; ++i, ++sectionHeader) {
                DWORD sectionStart = sectionHeader->VirtualAddress;
                DWORD sectionEnd = sectionStart + sectionHeader->Misc.VirtualSize;

                if (rva >= sectionStart && rva < sectionEnd) {
                    DWORD offset = rva - sectionStart + sectionHeader->PointerToRawData;
                    // Basic bounds check to ensure the offset is within the file
                    if (offset < fileBuffer.size()) {
                        return offset;
                    }
                }
            }
            return 0;
        }

        // CalculateEntropy Implementation
        double CalculateEntropy(const unsigned char* data, size_t size) {
            if (!data || size == 0) {
                return 0.0;
            }

            std::vector<long long> counts(256, 0);
            for (size_t i = 0; i < size; ++i) {
                counts[data[i]]++;
            }

            double entropy = 0.0;
            for (int i = 0; i < 256; ++i) {
                if (counts[i] > 0) {
                    double probability = static_cast<double>(counts[i]) / size;
                    entropy -= probability * log2(probability);
                }
            }
            return entropy;
        }

    } // namespace Internal
} // namespace PeAnalysis