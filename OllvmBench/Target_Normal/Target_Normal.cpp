#include "../BenchCore.h" // Adjust path if needed, or rely on Shared Project include path
#include <iostream>

int main() {
    std::cout << "Starting NORMAL Baseline Benchmark...\n\n";

    // We pass the label string to identify the run in the output
    BenchSuite::RunAll("Target: NORMAL (MSVC / No Obfuscation)");

    std::cout << "Press ENTER to exit...";
    std::cin.get();
    return 0;
}