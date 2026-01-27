#include "SimplePeParser.h"
#include <fstream>
#include <filesystem>

namespace fs = std::filesystem;



constexpr uint16_t E_LFANEW_OFFSET          = 0x3C;
constexpr uint16_t IMAGE_FILE_DLL           = 0x2000;
constexpr uint16_t MAGIC_BYTE_MZ            = 0x5A4D;      // 'MZ'
constexpr uint32_t PE_SIGNATURE             = 0x00004550;   // 'PE\0\0' (little endian)
constexpr uint16_t CHARACTERISTICS_OFFSET   = 18;



bool PEParser::IsDll(const fs::path& filePath) noexcept {
    namespace fs = std::filesystem;
    std::error_code ec;

    // 0. Check existence and get file size
    auto fileSize = fs::file_size(filePath, ec);
    if (ec || fileSize < sizeof(uint16_t) + 4) {
        return false;
    }

    std::ifstream file{ filePath, std::ios::binary, _SH_DENYWR };
    file.exceptions(std::ios::goodbit);
    if (!file.is_open()) {
        return false;
    }

    // 1. Read DOS signature and e_lfanew
    uint16_t dosSig = 0;
    file.read(reinterpret_cast<char*>(&dosSig), sizeof(dosSig));
    if (!file || dosSig != MAGIC_BYTE_MZ) {
        return false;
    }

    // Seek to e_lfanew offset field
    file.seekg(E_LFANEW_OFFSET, std::ios::beg);
    if (!file) {
        return false;
    }

    uint32_t peOffset = 0;
    file.read(reinterpret_cast<char*>(&peOffset), sizeof(peOffset));
    if (!file) {
        return false;
    }

    // Validate peOffset and needed header size
    //Read FileHeader.Characteristics(2 bytes) at offset 18 in FileHeader
    const uint32_t minSize = peOffset + sizeof(uint32_t) + CHARACTERISTICS_OFFSET + sizeof(uint16_t);
    if (peOffset > fileSize || minSize > fileSize) {
        return false;
    }

    // 2. Read PE signature
    file.seekg(peOffset, std::ios::beg);
    uint32_t peSig = 0;
    file.read(reinterpret_cast<char*>(&peSig), sizeof(peSig));
    if (!file || peSig != PE_SIGNATURE) { 
        return false;
    }

    // 3. Skip IMAGE_FILE_HEADER fields up to 'Characteristics' (18 bytes)
    file.seekg(CHARACTERISTICS_OFFSET, std::ios::cur);
    if (!file) {
        return false;
    }

    // 4. Read Characteristics
    uint16_t characteristics = 0;
    file.read(reinterpret_cast<char*>(&characteristics), sizeof(characteristics));
    if (!file) {
        return false;
    }

    // 5. Check IMAGE_FILE_DLL flag
    return (characteristics & IMAGE_FILE_DLL) != 0;
}