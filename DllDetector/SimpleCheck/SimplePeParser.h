#ifndef PE_PARSER_H
#define PE_PARSER_H

#include <cstdint>
#include <filesystem>
#include <string>

namespace fs = std::filesystem;

 /**
  * @namespace PEParser
  * @brief Utilities for parsing Portable Executable (PE) files.
  */
namespace PEParser {

	/**
	 * @brief Determines whether a given file is a Windows DLL by inspecting its PE header.
	 *
	 * Opens the file in binary mode, reads the DOS header to find the PE header,
	 * validates the PE signature, then checks the IMAGE_FILE_HEADER.Characteristics
	 * for the IMAGE_FILE_DLL flag (0x2000).
	 *
	 * @param filePath Path to the file to analyze.
	 * @return true if the file is a valid PE DLL; false if not a DLL or on any error (file not found, invalid PE, etc.).
	 * @note This function is marked noexcept and will catch and suppress all exceptions internally.
	 */
	bool IsDll(const fs::path& filePath) noexcept;

} // namespace PEParser

#endif // PE_PARSER_H
