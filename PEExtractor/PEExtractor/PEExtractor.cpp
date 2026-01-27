#include <iostream>
#include <fstream>
#include <filesystem>
#include <vector>
#include <string>

// Include your library header
#include "PEFeatureExtractor.h"

namespace fs = std::filesystem;

// Helper to sanitize CSV entries
std::string sanitize(const std::string& input) {
    if (input.find(',') != std::string::npos) {
        return "\"" + input + "\"";
    }
    return input;
}

int main(int argc, char* argv[]) {
    // 1. Argument Validation
    if (argc < 2) {
        std::cerr << "Usage: " << argv[0] << " <dataset_directory_path> [output_csv_path]" << std::endl;
        std::cerr << "Example: " << argv[0] << " \"D:\\dev\\source\\MalWare Files\" dataset.csv" << std::endl;
        return 1;
    }

    fs::path root_path = argv[1];
    std::string output_csv = (argc >= 3) ? argv[2] : "malware_dataset.csv";

    // 2. Check if directory exists
    if (!fs::exists(root_path) || !fs::is_directory(root_path)) {
        std::cerr << "[!] Error: The path '" << root_path.string() << "' is not a valid directory." << std::endl;
        return 1;
    }

    std::cout << "[*] Starting Dataset Generation..." << std::endl;
    std::cout << "[*] Root Directory: " << root_path.string() << std::endl;
    std::cout << "[*] Output File:    " << output_csv << std::endl;

    std::ofstream csv_file(output_csv);
    if (!csv_file.is_open()) {
        std::cerr << "[!] Error: Could not create output file " << output_csv << std::endl;
        return 1;
    }

    bool headers_written = false;
    int success_count = 0;
    int fail_count = 0;

    // 3. Iterate over the directories
    for (const auto& entry : fs::directory_iterator(root_path)) {
        if (!entry.is_directory()) continue;

        // The folder name becomes the label
        std::string label = entry.path().filename().string();
        std::cout << "Processing category: " << label << "..." << std::endl;

        // Iterate over files inside the category folder
        for (const auto& file_entry : fs::directory_iterator(entry.path())) {
            if (!file_entry.is_regular_file()) continue;

            try {
                // Pass the path as wstring to your library
                Radar::PEFeatureExtractor extractor(file_entry.path().wstring());

                std::vector<float> features = extractor.generate_feature_vector();

                if (features.empty()) {
                    fail_count++;
                    continue;
                }

                // Write Header Row (only once)
                if (!headers_written) {
                    for (size_t i = 0; i < features.size(); ++i) {
                        csv_file << "feat_" << i << ",";
                    }
                    csv_file << "label\n";
                    headers_written = true;
                }

                // Write Feature Data
                for (const auto& val : features) {
                    csv_file << val << ",";
                }

                // Write Label
                csv_file << sanitize(label) << "\n";

                success_count++;

                if (success_count % 100 == 0) {
                    std::cout << "\r   Processed " << success_count << " samples..." << std::flush;
                }

            }
            catch (const std::exception& e) {
                std::cerr << "\n[!] Exception on file " << file_entry.path().string() << ": " << e.what() << std::endl;
                fail_count++;
            }
        }
        std::cout << std::endl;
    }

    csv_file.close();

    std::cout << "------------------------------------------------" << std::endl;
    std::cout << "[*] Dataset generation complete." << std::endl;
    std::cout << "[+] Successfully processed: " << success_count << std::endl;
    std::cout << "[-] Failed/Skipped:         " << fail_count << std::endl;

    return 0;
}